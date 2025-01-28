package sync

import (
	"context"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/data"
	syncData "github.com/hgajjar/toolbox/data/sync"
)

func castToMappingEntities[T MappingInterface](entities []T) []MappingInterface {
	mappings := []MappingInterface{}
	for _, e := range entities {
		mappings = append(mappings, e)
	}

	return mappings
}

func castChannelToSyncEntity[T EntityInterface](from <-chan T) <-chan EntityInterface {
	to := make(chan EntityInterface)
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

func NewPlugin(repo *syncData.Repository, config *config.SyncEntity) *Sync {
	return &Sync{
		repo:   repo,
		config: config,
	}
}

func (a *Sync) GetData(ctx context.Context, filter data.Filter) (<-chan EntityInterface, error) {
	dataCh, err := a.repo.GetData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castChannelToSyncEntity(dataCh), nil
}

func (a *Sync) GetResourceName() string {
	return a.config.Resource
}

func (a *Sync) GetMappings() []MappingInterface {
	mappings := a.config.Mappings

	return castToMappingEntities(mappings)
}

func (a *Sync) GetQueueName() string {
	return a.config.QueueGroup
}
