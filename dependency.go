package ectoinject

import (
	"context"
	"fmt"
	"reflect"

	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/lifecycles"
)

type InstanceFunc func(context.Context, *EctoContainer) (any, error)

type EctoDependency struct {
	dependencyType      reflect.Type
	dependencyName      string
	dependencyValueType reflect.Type
	lifecycle           string
	value               reflect.Value
	getInstanceFunc     InstanceFunc
	constructor         reflect.Method
	constructorName     string
}

func (d *EctoDependency) SetValue(instance any) error {
	// get casted value
	val, err := ectoreflect.CastType(d.dependencyValueType, instance)
	if err != nil {
		return fmt.Errorf("failed to cast value for dependency '%s': %w", d.dependencyName, err)
	}

	d.value = val

	return nil
}

func (d *EctoDependency) GetValue() reflect.Value {
	return d.value
}

func (d *EctoDependency) HasValue() bool {
	return d.value != (reflect.Value{})
}

func (d *EctoDependency) HasConstructor() bool {
	return d.constructor != (reflect.Method{})
}
func (d *EctoDependency) GetConstructor() reflect.Method {
	return d.constructor
}

func (d *EctoDependency) GetInstanceFunc() InstanceFunc {
	return d.getInstanceFunc
}

func (d *EctoDependency) GetDependencyType() reflect.Type {
	return d.dependencyType
}

func (d *EctoDependency) GetDependencyName() string {
	return d.dependencyName
}

func (d *EctoDependency) GetDependencyValueType() reflect.Type {
	return d.dependencyValueType
}

func (d *EctoDependency) GetLifecycle() string {
	return d.lifecycle
}

func NewDependency[TType any](name, lifecycle, constructorName string, valueType reflect.Type, getInstanceFunc InstanceFunc) (Dependency, error) {
	dep := &EctoDependency{}
	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetIntefaceName[TType]()
	}

	// Ensure the lifecycle is valid
	if !lifecycles.IsValid(lifecycle) {
		return dep, fmt.Errorf("invalid lifecycle '%s' must be one of %v", lifecycle, lifecycles.Lifecycles)
	}

	dep.dependencyType = reflect.TypeOf((*TType)(nil)).Elem()
	dep.dependencyName = name
	dep.lifecycle = lifecycle
	dep.constructorName = constructorName
	dep.dependencyValueType = valueType

	if constructorName != "" {
		constructor, ok := ectoreflect.GetMethodByName(dep.GetDependencyValueType(), constructorName)
		if ok {
			dep.constructor = constructor
		}
	}

	if getInstanceFunc != nil {
		dep.getInstanceFunc = getInstanceFunc
	}

	return dep, nil
}

func NewDependencyWithInsance[TType any](name string, instance any) Dependency {
	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetIntefaceName[TType]()
	}

	return &EctoDependency{
		dependencyType:      reflect.TypeOf((*TType)(nil)).Elem(),
		dependencyName:      name,
		dependencyValueType: reflect.TypeOf(instance),
		lifecycle:           lifecycles.Singleton,
		value:               reflect.ValueOf(instance),
	}
}

func NewDependencyValue(name, lifecycle string, v any) (Dependency, error) {
	// Ensure the lifecycle is valid
	if !lifecycles.IsValid(lifecycle) {
		return &EctoDependency{}, fmt.Errorf("invalid lifecycle '%s' must be one of %v", lifecycle, lifecycles.Lifecycles)
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
		return &EctoDependency{}, fmt.Errorf("type '%s' is not a struct", t.Name())
	}

	if name == "" {
		// if a name is not provided, use the name of the interface
		name = ectoreflect.GetReflectTypeName(t)
	}

	return &EctoDependency{
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

	return &EctoDependency{
		dependencyType:  reflect.TypeOf((*TType)(nil)).Elem(),
		dependencyName:  name,
		lifecycle:       lifecycles.Singleton,
		getInstanceFunc: f,
	}
}
