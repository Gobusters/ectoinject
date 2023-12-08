package scope

import (
	"context"
	"strconv"
	"time"

	"github.com/Gobusters/ectoinject/dependency"
)

var contextScopedContainerIDKey = "ectoinject-dependency-Scoped-container-id"
var scopedCache = map[string]scopeCacheItem{}

type scopeCacheItem struct {
	scopedID  string
	instances map[string]dependency.Dependency
}

func AddScopedDependency(scopedID string, dep dependency.Dependency) {
	if _, ok := scopedCache[scopedID]; !ok {
		scopedCache[scopedID] = scopeCacheItem{
			scopedID:  scopedID,
			instances: map[string]dependency.Dependency{},
		}
	}

	scopedCache[scopedID].instances[dep.GetName()] = dep
}

func GetScopedDependency(scopedID, dependencyName string) (dependency.Dependency, bool) {
	if scopedID == "" {
		return nil, false
	}
	if _, ok := scopedCache[scopedID]; !ok {
		return nil, false
	}

	instance, ok := scopedCache[scopedID].instances[dependencyName]

	return instance, ok
}

func RemoveScopedCache(scopedID string) {
	delete(scopedCache, scopedID)
}

func RemoveScopedDependency(scopedID, dependencyName string) {
	if _, ok := scopedCache[scopedID]; !ok {
		return
	}

	delete(scopedCache[scopedID].instances, dependencyName)
}

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// scopeContext scopes the context to for use with scoped dependencies. To prevent memory leaks, the context must be canceled or UnscopeContext must be called when the scope is finished
// ctx: The context to scope
func ScopeContext(ctx context.Context) context.Context {
	// create a new Scoped container id
	id := generateID()

	// set the Scoped container id in the context
	ctx = context.WithValue(ctx, contextScopedContainerIDKey, id)

	// listen for the context to be done
	go func() {
		<-ctx.Done()
		// remove the Scoped container from the cache
		RemoveScopedCache(id)
	}()

	return ctx
}

// unscopeContext unscopes the context from a scoped dependency. This releases the cache for the scope and should be called when the scope is finished
// ctx: The context to unscope
func UnscopeContext(ctx context.Context) context.Context {
	id := GetScopedID(ctx)
	// remove the Scoped container id from the context
	ctx = context.WithValue(ctx, contextScopedContainerIDKey, "")

	// remove the Scoped container from the cache
	RemoveScopedCache(id)

	return ctx
}

func GetScopedID(ctx context.Context) string {
	id, _ := ctx.Value(contextScopedContainerIDKey).(string)
	return id
}
