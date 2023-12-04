package cache

var scopedCache = map[string]scopeCacheItem{}

type scopeCacheItem struct {
	scopedID  string
	instances map[string]any
}

func AddScopedInstance(scopedID, instanceName string, instance any) {
	if _, ok := scopedCache[scopedID]; !ok {
		scopedCache[scopedID] = scopeCacheItem{
			scopedID:  scopedID,
			instances: map[string]any{},
		}
	}

	scopedCache[scopedID].instances[instanceName] = instance
}

func GetScopedInstance(scopedID, instanceName string) (any, bool) {
	if scopedID == "" {
		return nil, false
	}
	if _, ok := scopedCache[scopedID]; !ok {
		return nil, false
	}

	instance, ok := scopedCache[scopedID].instances[instanceName]

	return instance, ok
}

func RemoveScopedCache(scopedID string) {
	delete(scopedCache, scopedID)
}

func RemoveScopedInstance(scopedID, instanceName string) {
	if _, ok := scopedCache[scopedID]; !ok {
		return
	}

	delete(scopedCache[scopedID].instances, instanceName)
}
