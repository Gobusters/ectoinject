package scope

import (
	"context"

	"github.com/Gobusters/ectoinject/dependency"
)

var contextScopedContainerIDKey = "ectoinject-dependency-scoped-container"

func AddScopedDependency(ctx context.Context, dep dependency.Dependency) context.Context {
	// get the scoped cache from the context
	cache, ok := ctx.Value(contextScopedContainerIDKey).(map[string]dependency.Dependency)
	if !ok {
		cache = make(map[string]dependency.Dependency)
	}

	// add the dependency to the cache
	cache[dep.GetName()] = dep

	// add the cache to the context
	return context.WithValue(ctx, contextScopedContainerIDKey, cache)
}

func GetScopedDependency(ctx context.Context, dependencyName string) (dependency.Dependency, bool) {
	// get the scoped cache from the context
	cache, ok := ctx.Value(contextScopedContainerIDKey).(map[string]dependency.Dependency)
	if !ok {
		cache = make(map[string]dependency.Dependency)
	}

	// get the dependency from the cache
	dep, ok := cache[dependencyName]
	return dep, ok
}
