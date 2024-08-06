package plugin

import "queue-worker/sync"

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
