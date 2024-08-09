package plugin

import (
	"context"
	"queue-worker/config"
	"queue-worker/data"
	syncData "queue-worker/data/sync"
	"queue-worker/sync"
)

func castToMappingEntities[T sync.MappingInterface](entities []T) []sync.MappingInterface {
	mappings := []sync.MappingInterface{}
	for _, e := range entities {
		mappings = append(mappings, e)
	}

	return mappings
}

func castChannelToSyncEntity[T sync.EntityInterface](from <-chan T) <-chan sync.EntityInterface {
	to := make(chan sync.EntityInterface)
	go func() {
		var val T
		for {
			val = <-from
			to <- val
		}
	}()
	return to
}

type Sync struct {
	repo   *syncData.Repository
	config *config.SyncEntity
}

func New(repo *syncData.Repository, config *config.SyncEntity) *Sync {
	return &Sync{
		repo:   repo,
		config: config,
	}
}

func (a *Sync) GetData(ctx context.Context, filter data.Filter) (<-chan sync.EntityInterface, error) {
	dataCh, err := a.repo.GetData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castChannelToSyncEntity(dataCh), nil
}

func (a *Sync) GetResourceName() string {
	return a.config.Resource
}

func (a *Sync) GetMappings() []sync.MappingInterface {
	mappings := a.config.Mappings

	return castToMappingEntities(mappings)
}

func (a *Sync) GetQueueName() string {
	return a.config.QueueGroup
}
