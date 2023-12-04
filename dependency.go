package ectoinject

import (
	"context"
	"fmt"
	"reflect"

	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/lifecycles"
)

type InstanceFunc func(context.Context, *DIContainer) (any, error)

type Dependency struct {
	DependencyType      reflect.Type
	DependencyName      string
	DependencyValueType reflect.Type
	Lifecycle           string
	Instance            any
	GetInstanceFunc     InstanceFunc
}

func NewDependency[TType any, TValue any](name, lifecycle string) (Dependency, error) {
	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetIntefaceName[TType]()
	}

	// Ensure the lifecycle is valid
	if !lifecycles.IsValid(lifecycle) {
		return Dependency{}, fmt.Errorf("invalid lifecycle '%s' must be one of %v", lifecycle, lifecycles.Lifecycles)
	}

	return Dependency{
		DependencyType:      reflect.TypeOf((*TType)(nil)).Elem(),
		DependencyName:      name,
		DependencyValueType: reflect.TypeOf((*TValue)(nil)).Elem(),
		Lifecycle:           lifecycle,
	}, nil
}

func NewDependencyWithInsance[TType any](name string, instance any) Dependency {
	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetIntefaceName[TType]()
	}

	return Dependency{
		DependencyType:      reflect.TypeOf((*TType)(nil)).Elem(),
		DependencyName:      name,
		DependencyValueType: reflect.TypeOf(instance),
		Lifecycle:           lifecycles.Singleton,
		Instance:            instance,
	}
}

func NewDependencyValue(name, lifecycle string, v any) (Dependency, error) {
	// Ensure the lifecycle is valid
	if !lifecycles.IsValid(lifecycle) {
		return Dependency{}, fmt.Errorf("invalid lifecycle '%s' must be one of %v", lifecycle, lifecycles.Lifecycles)
	}

	var t reflect.Type
	// if v is reflect.Type, use the type directly
	if _, ok := v.(reflect.Type); ok {
		t = v.(reflect.Type)
	} else {
		// otherwise, get the type of v
		t = reflect.TypeOf(v)
	}

	// Ensure t is a struct type
	if t.Kind() != reflect.Struct {
		return Dependency{}, fmt.Errorf("type '%s' is not a struct", t.Name())
	}

	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetReflectTypeName(t)
	}

	return Dependency{
		DependencyType:      t,
		DependencyName:      name,
		DependencyValueType: t,
		Lifecycle:           lifecycle,
	}, nil
}

func NewCustomFuncDependency[TType any](name string, f InstanceFunc) Dependency {
	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetIntefaceName[TType]()
	}

	return Dependency{
		DependencyType:  reflect.TypeOf((*TType)(nil)).Elem(),
		DependencyName:  name,
		Lifecycle:       lifecycles.Singleton,
		GetInstanceFunc: f,
	}
}
