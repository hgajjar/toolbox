package plugin

import (
	"context"
	"queue-worker/data"
	"queue-worker/data/availability"
	"queue-worker/sync"
)

type AvailabilityStorageSync struct {
	repo *availability.Repository
}

func NewAvailabilityStorageSync(repo *availability.Repository) *AvailabilityStorageSync {
	return &AvailabilityStorageSync{
		repo: repo,
	}
}

func (a *AvailabilityStorageSync) GetData(ctx context.Context, filter data.Filter) ([]sync.EntityInterface, error) {
	entities, err := a.repo.GetStorageData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castToSyncEntities(entities), nil
}

func (a *AvailabilityStorageSync) GetResourceName() string {
	return a.repo.GetResourceName()
}

func (a *AvailabilityStorageSync) GetMappings() []sync.MappingInterface {
	mappings := a.repo.GetMappings()

	return castToMappingEntities(mappings)
}

func (a *AvailabilityStorageSync) GetQueueName() string {
	return "sync.storage.availability"
}
