package scope

import (
	"context"

	"github.com/Gobusters/ectoinject/dependency"
)

type contextKey string

var contextScopedContainerIDKey = contextKey("ectoinject-dependency-scoped-container")

// AddScopedDependency adds a scoped dependency to the context. This allows for scoped caching of dependencies on the context
// ctx: The context to add the scoped dependency to
// dep: The dependency to add
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

// GetScopedDependency gets a scoped dependency from the context. Returns the dependency and a bool indicating if the dependency was found
// ctx: The context to get the scoped dependency from
// dependencyName: The name of the dependency to get
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
