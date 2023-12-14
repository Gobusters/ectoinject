package ectoinject

import (
	"context"

	"github.com/Gobusters/ectoinject/ectocontainer"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
)

// GetFromContainer gets a dependency from the container. Returns the dependency and an error.
// T: The type of the dependency
// containerID: The id of the container to get the dependency from
func GetFromContainer[T any](containerID string) (T, error) {
	ctx := context.Background()
	ctx, err := SetActiveContainer(ctx, containerID)
	if err != nil {
		var zero T
		return zero, err
	}

	_, val, err := GetNamedDependency[T](ctx, "")
	return val, err
}

// Get gets a dependency from the default container. Returns the dependency and an error.
// T: The type of the dependency
func Get[T any]() (T, error) {
	ctx := context.Background()
	_, val, err := GetNamedDependency[T](ctx, "")
	return val, err
}

// GetContext gets a dependency from the container. Returns a context with scoped dependencies caching, the dependency, and an error.
// The context is used to provide scoped caching of dependencies.
// T: The type of the dependency
// ctx: The context to use. To use a non-default container, use SetActiveContainer
func GetContext[T any](ctx context.Context) (context.Context, T, error) {
	return GetNamedDependency[T](ctx, "")
}

// GetNamedDependency gets a dependency from the container by name
// T: The type of the dependency
// ctx: The context to use. To use a non-default container, use SetActiveContainer
// name: The name of the dependency. Unnamed dependencies will use `{module}.{type}` as the name
func GetNamedDependency[T any](ctx context.Context, name string) (context.Context, T, error) {
	var val T

	activeContainer, err := GetActiveContainer(ctx)
	if err != nil {
		return ctx, val, err
	}

	reflectName := ectoreflect.GetIntefaceName[T]()
	if name == "" {
		name = reflectName
	}

	return GetDependency[T](ctx, activeContainer, name)
}

// GetDependency gets a dependency from the container by name
// T: The type of the dependency
// ctx: The context to use. To use a non-default container, use SetActiveContainer
// container: The container to get the dependency from
// name: The name of the dependency. Unnamed dependencies will use `{module}.{type}` as the name
func GetDependency[T any](ctx context.Context, container ectocontainer.DIContainer, name string) (context.Context, T, error) {
	var val T

	ctx, depValue, err := container.Get(ctx, name)
	if err != nil {
		return ctx, val, err
	}

	val, err = ectoreflect.Cast[T](depValue)
	if err != nil {
		return ctx, val, err
	}

	return ctx, val, nil
}
