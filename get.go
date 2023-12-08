package ectoinject

import (
	"context"

	"github.com/Gobusters/ectoinject/internal/container"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
)

// GetDependency gets a dependency from the container
// T: The type of the dependency
// ctx: The context to use. If you're using a scoped dependency, you must use a scoped context from ScopeContext
func GetDependency[T any](ctx context.Context) (T, error) {
	var val T

	activeContainer, err := GetActiveContainer(ctx)
	if err != nil {
		return val, err
	}

	reflectName := ectoreflect.GetIntefaceName[T]()

	// if the dependency is DIContainer, return the container
	if reflectName == ectoreflect.GetIntefaceName[DIContainer]() {
		id := defaultContainerID
		depContainer := container.GetContainer(id)
		return ectoreflect.Cast[T](depContainer)
	}

	depValue, err := activeContainer.Get(ctx, reflectName)
	if err != nil {
		return val, err
	}

	val, err = ectoreflect.Cast[T](depValue)
	if err != nil {
		return val, err
	}

	return val, nil
}

// GetDependency gets a dependency from the container
// T: The type of the dependency
// ctx: The context to use. If you're using a scoped dependency, you must use a scoped context from ScopeContext
func GetNamedDependency[T any](ctx context.Context, name string) (T, error) {
	var val T

	activeContainer, err := GetActiveContainer(ctx)
	if err != nil {
		return val, err
	}

	reflectName := ectoreflect.GetIntefaceName[T]()
	depName := name
	if depName == "" {
		depName = reflectName
	}

	// if the dependency is DIContainer, return the container
	if reflectName == ectoreflect.GetIntefaceName[DIContainer]() {
		id := defaultContainerID
		depContainer := container.GetContainer(id)
		return ectoreflect.Cast[T](depContainer)
	}

	depValue, err := activeContainer.Get(ctx, depName)
	if err != nil {
		return val, err
	}

	val, err = ectoreflect.Cast[T](depValue)
	if err != nil {
		return val, err
	}

	return val, nil
}
