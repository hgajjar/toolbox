package plugin

import (
	"context"
	"queue-worker/data"
	"queue-worker/data/category"
	"queue-worker/sync"
)

type CategoryImageStorageSync struct {
	repo *category.Repository
}

func NewCategoryImageStorageSync(repo *category.Repository) *CategoryImageStorageSync {
	return &CategoryImageStorageSync{
		repo: repo,
	}
}

func (p *CategoryImageStorageSync) GetData(ctx context.Context, filter data.Filter) (<-chan sync.EntityInterface, error) {
	entities, err := p.repo.GetCategoryImageStorageData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castChannelToSyncEntity(entities), nil
}

func (p *CategoryImageStorageSync) GetResourceName() string {
	return p.repo.GetCateogryImageResourceName()
}

func (p *CategoryImageStorageSync) GetMappings() []sync.MappingInterface {
	mappings := p.repo.GetCategoryImageMappings()

	return castToMappingEntities(mappings)
}

func (p *CategoryImageStorageSync) GetQueueName() string {
	return "sync.storage.category"
}
