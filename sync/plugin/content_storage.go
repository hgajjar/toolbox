package plugin

import (
	"context"
	"queue-worker/data"
	"queue-worker/data/content"
	"queue-worker/sync"
)

type ContentStorageSync struct {
	repo *content.Repository
}

func NewContentStorageSync(repo *content.Repository) *ContentStorageSync {
	return &ContentStorageSync{
		repo: repo,
	}
}

func (p *ContentStorageSync) GetData(ctx context.Context, filter data.Filter) (<-chan sync.EntityInterface, error) {
	dataCh, err := p.repo.GetContentStorageData(ctx, filter)
	if err != nil {
		return nil, err
	}

	return castChannelToSyncEntity(dataCh), nil
}

func (p *ContentStorageSync) GetResourceName() string {
	return p.repo.GetContentResourceName()
}

func (p *ContentStorageSync) GetMappings() []sync.MappingInterface {
	mappings := p.repo.GetContentMappings()

	return castToMappingEntities(mappings)
}

func (p *ContentStorageSync) GetQueueName() string {
	return "sync.storage.content"
}
