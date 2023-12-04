package ectoinject

import (
	"fmt"

	"github.com/Gobusters/ectoinject/lifecycles"
)

// RegisterSingleton registers a singleton dependency in the container. Singleton dependencies are cached for the lifetime of the application
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
func RegisterSingleton[TType any, TValue any](container *DIContainer) error {
	return RegisterDependency[TType, TValue](container, "", lifecycles.Singleton)
}

// RegisterScoped registers a scoped dependency in the container. Scoped dependencies are cached for the lifetime of the scope
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
func RegisterScoped[TType any, TValue any](container *DIContainer) error {
	return RegisterDependency[TType, TValue](container, "", lifecycles.Scoped)
}

// RegisterTransient registers a transient dependency in the container. Transient dependencies are not cached
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
func RegisterTransient[TType any, TValue any](container *DIContainer) error {
	return RegisterDependency[TType, TValue](container, "", lifecycles.Transient)
}

// RegisterNamedSingleton registers a singleton dependency in the container. Singleton dependencies are cached for the lifetime of the application
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// name: The name of the dependency
func RegisterNamedSingleton[TType any, TValue any](container *DIContainer, name string) error {
	return RegisterDependency[TType, TValue](container, name, lifecycles.Singleton)
}

// RegisterNamedScoped registers a scoped dependency in the container. Scoped dependencies are cached for the lifetime of the scope
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// name: The name of the dependency
func RegisterNamedScoped[TType any, TValue any](container *DIContainer, name string) error {
	return RegisterDependency[TType, TValue](container, name, lifecycles.Scoped)
}

// RegisterNamedTransient registers a transient dependency in the container. Transient dependencies are not cached
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// name: The name of the dependency
func RegisterNamedTransient[TType any, TValue any](container *DIContainer, name string) error {
	return RegisterDependency[TType, TValue](container, name, lifecycles.Transient)
}

// RegisterInstance registers an instance in the container. Instances are treated as singletons
// TType: The type of the dependency
// container: The container to register the dependency in
// instance: The instance to register
func RegisterInstance[TType any](container *DIContainer, instance any) error {
	return RegisterNamedInstance[TType](container, "", instance)
}

// RegisterNamedInstance registers an instance in the container. Instances are treated as singletons
// TType: The type of the dependency
// container: The container to register the dependency in
// name: The name of the dependency
// instance: The instance to register
func RegisterNamedInstance[TType any](container *DIContainer, name string, instance any) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	dep := NewDependencyWithInsance[TType](name, instance)

	container.container[dep.DependencyName] = dep

	return nil
}

// RegisterDependency registers a dependency in the container
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// lifecycle: The lifecycle of the dependency
func RegisterDependency[TType any, TValue any](container *DIContainer, name, lifecycle string) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	dep, err := NewDependency[TType, TValue](name, lifecycle)
	if err != nil {
		return err
	}

	container.container[dep.DependencyName] = dep

	return nil
}

// RegisterNamedCustomDependencyFunc registers a custom dependency handler in the container
// TType: The type of the dependency
// container: The container to register the dependency in
// name: The name of the dependency
// f: The function to call to get the instance
func RegisterNamedCustomDependencyFunc[TType any](container *DIContainer, name string, f InstanceFunc) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	dep := NewCustomFuncDependency[TType](name, f)

	dep.GetInstanceFunc = f

	container.container[dep.DependencyName] = dep

	return nil
}

// RegisterCustomDependencyFunc registers a custom dependency handler in the container
// TType: The type of the dependency
// container: The container to register the dependency in
// f: The function to call to get the instance
func RegisterCustomDependencyFunc[TType any](container *DIContainer, f InstanceFunc) error {
	return RegisterNamedCustomDependencyFunc[TType](container, "", f)
}

// RegisterValue registers a value in the container
// container: The container to register the dependency in
// name: The name of the dependency
// lifecycle: The lifecycle of the dependency
// v: The dependency to register
func RegisterValue(container *DIContainer, name, lifecycle string, v any) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	dep, err := NewDependencyValue(name, lifecycle, v)
	if err != nil {
		return err
	}

	container.container[dep.DependencyName] = dep

	return nil
}
