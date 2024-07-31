package plugin

import (
	"context"
	"queue-worker/data"
	"queue-worker/data/category"
	"queue-worker/sync"
)

type CategoryNodeStorageSync struct {
	repo *category.Repository
}

func NewCategoryNodeStorageSync(repo *category.Repository) *CategoryNodeStorageSync {
	return &CategoryNodeStorageSync{
		repo: repo,
	}
}

func (p *CategoryNodeStorageSync) GetData(ctx context.Context, filter data.Filter) ([]sync.EntityInterface, error) {
	entities, err := p.repo.GetCategoryNodeStorageData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castToSyncEntities(entities), nil
}

func (p *CategoryNodeStorageSync) GetResourceName() string {
	return p.repo.GetCateogryNodeResourceName()
}

func (p *CategoryNodeStorageSync) GetMappings() []sync.MappingInterface {
	mappings := p.repo.GetCategoryNodeMappings()

	return castToMappingEntities(mappings)
}

func (p *CategoryNodeStorageSync) GetQueueName() string {
	return "sync.storage.category"
}
