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
	dependencyType      reflect.Type
	dependencyName      string
	dependencyValueType reflect.Type
	lifecycle           string
	instance            any
	getInstanceFunc     InstanceFunc
	constructor         reflect.Method
}

func (d Dependency) hasConstructor() bool {
	return d.constructor != (reflect.Method{})
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
		dependencyType:      reflect.TypeOf((*TType)(nil)).Elem(),
		dependencyName:      name,
		dependencyValueType: reflect.TypeOf((*TValue)(nil)).Elem(),
		lifecycle:           lifecycle,
	}, nil
}

func NewDependencyWithInsance[TType any](name string, instance any) Dependency {
	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetIntefaceName[TType]()
	}

	return Dependency{
		dependencyType:      reflect.TypeOf((*TType)(nil)).Elem(),
		dependencyName:      name,
		dependencyValueType: reflect.TypeOf(instance),
		lifecycle:           lifecycles.Singleton,
		instance:            instance,
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
		dependencyType:      t,
		dependencyName:      name,
		dependencyValueType: t,
		lifecycle:           lifecycle,
	}, nil
}

func NewCustomFuncDependency[TType any](name string, f InstanceFunc) Dependency {
	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetIntefaceName[TType]()
	}

	return Dependency{
		dependencyType:  reflect.TypeOf((*TType)(nil)).Elem(),
		dependencyName:  name,
		lifecycle:       lifecycles.Singleton,
		getInstanceFunc: f,
	}
}
