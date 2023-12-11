package ectoinject

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Gobusters/ectoinject/internal/dependency"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/lifecycles"
)

// RegisterSingleton registers a singleton dependency in the container. Singleton dependencies are cached for the lifetime of the application
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// names: (optional) The names of the dependency
func RegisterSingleton[TType any, TValue any](container DIContainer, names ...string) error {
	return RegisterDependency[TType, TValue](container, lifecycles.Singleton, names...)
}

// RegisterScoped registers a scoped dependency in the container. Scoped dependencies are cached for the lifetime of the scope
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// names: (optional) The names of the dependency
func RegisterScoped[TType any, TValue any](container DIContainer, names ...string) error {
	return RegisterDependency[TType, TValue](container, lifecycles.Scoped, names...)
}

// RegisterTransient registers a transient dependency in the container. Transient dependencies are not cached
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// names: (optional) The names of the dependency
func RegisterTransient[TType any, TValue any](container DIContainer, names ...string) error {
	return RegisterDependency[TType, TValue](container, lifecycles.Transient, names...)
}

// RegisterInstanceFunc registers a custom instance function in the container
// TType: The type of the dependency
// container: The container to register the dependency in
// lifecycle: The lifecycle of the dependency. Must be one of transient, scoped, or singleton
// getInstanceFunc: a function that returns the instance
// names: (optional) The names of the dependency
func RegisterInstanceFunc[TType any](container DIContainer, lifecycle string, getInstanceFunc func(context.Context) (any, error), names ...string) error {
	if len(names) == 0 {
		names = []string{""}
	}
	valueType := reflect.TypeOf((*TType)(nil)).Elem()
	for _, name := range names {
		// create a new dependency
		dep, err := dependency.NewDependency[TType](name, lifecycle, container.GetConstructorFuncName(), valueType, getInstanceFunc)
		if err != nil {
			return err
		}

		// add the dependency to the container
		container.AddDependency(dep)
	}
	return nil
}

// RegisterInstance registers an instance in the container. Instances are treated as singletons
// TType: The type of the dependency
// container: The container to register the dependency in
// instance: The instance to register
// names: (optional) The names of the dependency
func RegisterInstance[TType any](container DIContainer, instance any, names ...string) error {
	getInstanceFunc := func(context.Context) (any, error) {
		return instance, nil
	}

	return RegisterInstanceFunc[TType](container, lifecycles.Singleton, getInstanceFunc, names...)
}

// RegisterDependency registers a dependency in the container
// TType: The type of the dependency
// TValue: The implementation of the dependency
// container: The container to register the dependency in
// lifecycle: The lifecycle of the dependency
func RegisterDependency[TType any, TValue any](container DIContainer, lifecycle string, names ...string) error {
	// ensure the TValue is a Struct
	valueType := reflect.TypeOf((*TValue)(nil)).Elem()
	if valueType.Kind() != reflect.Struct {
		return fmt.Errorf("dependency '%s' has type '%s' which is not a struct. Please register the dependnecy using RegisterInstance or RegisterInstanceFunc", valueType.Name(), valueType.Name())
	}

	// Ensure
	if !ectoreflect.SameType[TType, TValue]() {
		return fmt.Errorf("type '%s' is not assignable to '%s'", ectoreflect.GetReflectTypeName(valueType), ectoreflect.GetIntefaceName[TType]())
	}

	if len(names) == 0 {
		names = []string{""}
	}
	for _, name := range names {
		// create a new dependency
		dep, err := dependency.NewDependency[TType](name, lifecycle, container.GetConstructorFuncName(), valueType, nil)
		if err != nil {
			return err
		}

		// add the dependency to the container
		container.AddDependency(dep)
	}
	return nil
}
