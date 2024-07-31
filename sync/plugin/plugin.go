package plugin

import "queue-worker/sync"

func castToSyncEntities[T sync.EntityInterface](entities []T) []sync.EntityInterface {
	syncEntities := []sync.EntityInterface{}
	for _, e := range entities {
		syncEntities = append(syncEntities, e)
	}

	return syncEntities
}

func castToMappingEntities[T sync.MappingInterface](entities []T) []sync.MappingInterface {
	mappings := []sync.MappingInterface{}
	for _, e := range entities {
		mappings = append(mappings, e)
	}

	return mappings
}
