package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hgajjar/toolbox/data"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	rabbitMqBlockedConnWait = time.Millisecond * 200
	exportChunksize         = 10000
)

type Exporter struct {
	conn          *amqp.Connection
	connIsBlocked bool
	plugins       []SyncDataPluginInterface
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
	GenerateMappingKey(resourceName, source, sourceId string) string
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

func NewExporter(conn *amqp.Connection, plugins []SyncDataPluginInterface) *Exporter {
	return &Exporter{
		conn:    conn,
		plugins: plugins,
	}
}

func (e *Exporter) Export(ctx context.Context, IDs []int) error {
	e.listenRabbitMqNotifications(ctx)

	errs, ctx := errgroup.WithContext(ctx)

	for _, plugin := range e.plugins {
		errs.Go(func() error {
			return e.exportData(ctx, plugin, IDs)
		})
	}

	return errs.Wait()
}

func (e *Exporter) listenRabbitMqNotifications(ctx context.Context) {
	blockings := e.conn.NotifyBlocked(make(chan amqp.Blocking))

	go func(ctx context.Context, e *Exporter) {
		logger := zerolog.Ctx(ctx)
		for b := range blockings {
			if b.Active {
				e.connIsBlocked = true
				logger.Debug().Msgf("RabbitMQ connection blocked: %q", b.Reason)
			} else {
				e.connIsBlocked = false
				logger.Debug().Msgf("RabbitMQ connection unblocked")
			}
		}
	}(ctx, e)
}

func (e *Exporter) exportData(ctx context.Context, plugin SyncDataPluginInterface, IDs []int) error {
	offset := 0
	limit := exportChunksize

	rmqChannel, err := e.conn.Channel()
	if err != nil {
		zerolog.Ctx(ctx).Panic().Stack().Err(err).Msg("Failed to open a rabbitmq channel")
	}
	defer rmqChannel.Close()

	for {
		err, hasMore := e.exportDataChunk(ctx, plugin, rmqChannel, IDs, offset, limit)
		if err != nil {
			return err
		}
		if !hasMore {
			return nil
		}
		offset += limit
	}
}

func (e *Exporter) exportDataChunk(ctx context.Context, plugin SyncDataPluginInterface, rmqChannel *amqp.Channel, IDs []int, offset, limit int) (err error, hasMore bool) {
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

		var decodedVal any
		err = json.Unmarshal([]byte(entity.GetData()), &decodedVal)
		if err != nil {
			return err, false
		}

		m := syncMessage{
			message{
				entity.GetKey(),
				decodedVal,
				plugin.GetResourceName(),
				entity.GetStore(),
			},
		}
		j, err := json.Marshal(m)
		if err != nil {
			return err, false
		}

		err = e.publishMessage(ctx, rmqChannel, plugin.GetQueueName(), j)
		if err != nil {
			return err, false
		}

		if mappings := plugin.GetMappings(); mappings != nil {
			err = e.exportMappingData(ctx, rmqChannel, mappings, plugin.GetResourceName(), plugin.GetQueueName(), entity)
			if err != nil {
				return err, false
			}
		}

		hasMore = true
	}

	return
}

func (e *Exporter) publishMessage(parentCtx context.Context, rmqChannel *amqp.Channel, queueName string, body []byte) error {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	if e.connIsBlocked {
		e.waitUntilConnIsUnblocked(ctx)
	}
	err := rmqChannel.PublishWithContext(ctx,
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

func (e *Exporter) exportMappingData(ctx context.Context, rmqChannel *amqp.Channel, mappings []MappingInterface, resourceName string, queueName string, entity EntityInterface) error {

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

		sourceVal, ok := source.(string)
		if !ok {
			return fmt.Errorf("failed to parse mapping source value as string. Resource: %s, Data: %s", resourceName, entity.GetData())
		}

		key := entity.GenerateMappingKey(resourceName, mapping.GetSource(), sourceVal)
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

		err = e.publishMessage(ctx, rmqChannel, queueName, j)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Exporter) waitUntilConnIsUnblocked(ctx context.Context) {
	for {
		if !e.connIsBlocked {
			break
		}
		if e.conn.IsClosed() {
			zerolog.Ctx(ctx).Warn().Msg("RabbitMQ connection is closed unexpectedly")
		}
		time.Sleep(rabbitMqBlockedConnWait)
	}
}
