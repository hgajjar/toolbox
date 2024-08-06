package plugin

import (
	"context"
	"queue-worker/data"
	"queue-worker/data/category"
	"queue-worker/sync"
)

type CategoryTreeStorageSync struct {
	repo *category.Repository
}

func NewCategoryTreeStorageSync(repo *category.Repository) *CategoryTreeStorageSync {
	return &CategoryTreeStorageSync{
		repo: repo,
	}
}

func (p *CategoryTreeStorageSync) GetData(ctx context.Context, filter data.Filter) (<-chan sync.EntityInterface, error) {
	entities, err := p.repo.GetCategoryTreeStorageData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castChannelToSyncEntity(entities), nil
}

func (p *CategoryTreeStorageSync) GetResourceName() string {
	return p.repo.GetCateogryTreeResourceName()
}

func (p *CategoryTreeStorageSync) GetMappings() []sync.MappingInterface {
	mappings := p.repo.GetCategoryTreeMappings()

	return castToMappingEntities(mappings)
}

func (p *CategoryTreeStorageSync) GetQueueName() string {
	return "sync.storage.category"
}
