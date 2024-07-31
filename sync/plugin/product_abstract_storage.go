package plugin

import (
	"context"
	"queue-worker/data"
	"queue-worker/data/product"
	"queue-worker/sync"
)

type ProductAbstractStorageSync struct {
	repo *product.Repository
}

func NewProductAbstractStorageSync(repo *product.Repository) *ProductAbstractStorageSync {
	return &ProductAbstractStorageSync{
		repo: repo,
	}
}

func (p *ProductAbstractStorageSync) GetData(ctx context.Context, filter data.Filter) ([]sync.EntityInterface, error) {
	entities, err := p.repo.GetProductAbstractStorageData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castToSyncEntities(entities), nil
}

func (p *ProductAbstractStorageSync) GetResourceName() string {
	return p.repo.GetProductAbstractResourceName()
}

func (p *ProductAbstractStorageSync) GetMappings() []sync.MappingInterface {
	mappings := p.repo.GetProductAbstractMappings()

	return castToMappingEntities(mappings)
}

func (p *ProductAbstractStorageSync) GetQueueName() string {
	return "sync.storage.product"
}
