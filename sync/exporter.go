package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"queue-worker/data"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

type Exporter struct {
	channel *amqp.Channel
	plugins []SyncDataPluginInterface
}

type SyncDataPluginInterface interface {
	GetData(ctx context.Context, filter data.Filter) (<-chan EntityInterface, error)
	GetResourceName() string
	GetQueueName() string
	GetMappings() []MappingInterface
}

type EntityInterface interface {
	GetKey() string
	GetData() string
	GetStore() string
	GenerateMappingKey(source, sourceId string) string
	IsNil() bool
}

type MappingInterface interface {
	GetSource() string
	GetDestination() string
}

type message struct {
	Key      string `json:"key"`
	Value    any    `json:"value"`
	Resource string `json:"resource"`
	Store    string `json:"store,omitempty"`
}

type syncMessage struct {
	Write message `json:"write"`
}

type mappingValue map[string]any

func NewExporter(channel *amqp.Channel, plugins []SyncDataPluginInterface) *Exporter {
	return &Exporter{
		channel: channel,
		plugins: plugins,
	}
}

func (e *Exporter) Export(ctx context.Context, IDs []int) error {
	errs, ctx := errgroup.WithContext(ctx)

	for _, plugin := range e.plugins {
		errs.Go(func() error {
			return e.exportData(ctx, plugin, IDs)
		})
	}

	return errs.Wait()
}

func (e *Exporter) exportData(ctx context.Context, plugin SyncDataPluginInterface, IDs []int) error {
	chunksize := 10000
	offset := 0
	limit := chunksize

	for {
		err, hasMore := e.exportDataChunk(ctx, plugin, IDs, offset, limit)
		if err != nil {
			return err
		}
		if !hasMore {
			return nil
		}
		offset += chunksize
	}
}

func (e *Exporter) exportDataChunk(ctx context.Context, plugin SyncDataPluginInterface, IDs []int, offset, limit int) (err error, hasMore bool) {
	// Check if the context is expired.
	select {
	default:
	case <-ctx.Done():
		err = ctx.Err()
		return
	}

	syncEntityCh, err := plugin.GetData(ctx, data.NewFilter(offset, limit, IDs))
	if err != nil {
		return
	}

	for entity := range syncEntityCh {
		if entity.IsNil() {
			break
		}

		m := syncMessage{
			message{
				entity.GetKey(),
				entity.GetData(),
				plugin.GetResourceName(),
				entity.GetStore(),
			},
		}
		j, err := json.Marshal(m)
		if err != nil {
			return err, false
		}

		err = e.publishMessage(plugin.GetQueueName(), j)
		if err != nil {
			return err, false
		}

		if mappings := plugin.GetMappings(); mappings != nil {
			err = e.exportMappingData(mappings, plugin.GetResourceName(), plugin.GetQueueName(), entity)
			if err != nil {
				return err, false
			}
		}

		hasMore = true
	}

	return
}

func (e *Exporter) publishMessage(queueName string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := e.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})

	if err != nil {
		return fmt.Errorf("failed to publish a message. Error: %s", err.Error())
	}

	return nil
}

func (e *Exporter) exportMappingData(mappings []MappingInterface, resourceName string, queueName string, entity EntityInterface) error {

	for _, mapping := range mappings {

		var data map[string]any
		err := json.Unmarshal([]byte(entity.GetData()), &data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal entity data. Error: %s", err.Error())
		}

		source, ok := data[mapping.GetSource()]
		if !ok {
			return fmt.Errorf("entity data does not have key %s", mapping.GetSource())
		}
		destination, ok := data[mapping.GetDestination()]
		if !ok {
			return fmt.Errorf("entity data does not have key %s", mapping.GetDestination())
		}

		key := entity.GenerateMappingKey(mapping.GetSource(), source.(string))
		value := mappingValue{"id": destination, "_timestamp": time.Now().Unix()}

		m := syncMessage{
			message{
				Key:      key,
				Value:    value,
				Resource: resourceName,
			},
		}
		j, err := json.Marshal(m)
		if err != nil {
			return err
		}

		err = e.publishMessage(queueName, j)
		if err != nil {
			return err
		}
	}

	return nil
}
