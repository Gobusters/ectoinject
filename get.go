package ectoinject

import (
	"context"

	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
)

// GetDependency gets a dependency from the container
// T: The type of the dependency
// ctx: The context to use. To use a non-default container, use SetActiveContainer
func GetDependency[T any](ctx context.Context) (T, error) {
	return GetNamedDependency[T](ctx, "")
}

// GetNamedDependency gets a dependency from the container by name
// T: The type of the dependency
// ctx: The context to use. To use a non-default container, use SetActiveContainer
// name: The name of the dependency. Unnamed dependencies will use `{module}.{type}` as the name
func GetNamedDependency[T any](ctx context.Context, name string) (T, error) {
	var val T

	activeContainer, err := GetActiveContainer(ctx)
	if err != nil {
		return val, err
	}

	reflectName := ectoreflect.GetIntefaceName[T]()
	if name == "" {
		name = reflectName
	}

	depValue, err := activeContainer.Get(ctx, name)
	if err != nil {
		return val, err
	}

	val, err = ectoreflect.Cast[T](depValue)
	if err != nil {
		return val, err
	}

	return val, nil
}
