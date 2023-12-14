package dependency

import (
	"context"
	"fmt"
	"reflect"

	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/lifecycles"
)

type EctoDependency struct {
	dependencyType      reflect.Type
	dependencyName      string
	dependencyValueType reflect.Type
	lifecycle           string
	value               reflect.Value
	getInstanceFunc     func(context.Context) (any, error)
	constructor         reflect.Method
	constructorName     string
	instance            any
}

// SetValue sets the value of the dependency
func (d *EctoDependency) SetValue(v reflect.Value) error {
	d.value = v
	d.instance = ectoreflect.GetPointerOfValue(v)

	return nil
}

// GetValue gets the value of the dependency
func (d *EctoDependency) GetValue() reflect.Value {
	return d.value
}

// GetInstance gets the instance of the dependency
func (d *EctoDependency) GetInstance() (any, error) {
	if d.instance == nil {
		return nil, fmt.Errorf("dependency '%s' has no instance", d.dependencyName)
	}
	val, err := ectoreflect.CastType(d.dependencyType, d.instance)
	if err != nil {
		return nil, fmt.Errorf("failed to cast dependency '%s' to type '%s': %w", d.dependencyName, d.dependencyType.Name(), err)
	}

	d.instance = val.Interface()

	return d.instance, nil
}

// HasValue checks if the dependency has a value
func (d *EctoDependency) HasValue() bool {
	return d.value != (reflect.Value{})
}

// HasConstructor checks if the dependency has a constructor func
func (d *EctoDependency) HasConstructor() bool {
	return d.constructor != (reflect.Method{})
}

// GetConstructor gets the constructor func of the dependency
func (d *EctoDependency) GetConstructor() reflect.Method {
	return d.constructor
}

// GetInstanceFunc returns the custom instance func of the dependency
func (d *EctoDependency) GetInstanceFunc() func(context.Context) (any, error) {
	return d.getInstanceFunc
}

// GetDependencyType returns the type of the dependency
func (d *EctoDependency) GetDependencyType() reflect.Type {
	return d.dependencyType
}

// GetName returns the name of the dependency
func (d *EctoDependency) GetName() string {
	return d.dependencyName
}

// GetDependencyValueType gets the type of the dependency value
func (d *EctoDependency) GetDependencyValueType() reflect.Type {
	return d.dependencyValueType
}

// GetLifecycle returns the lifecycle of the dependency
func (d *EctoDependency) GetLifecycle() string {
	return d.lifecycle
}

// NewDependency creates a new EctoDependency
// TType: The type of the dependency
// name: The name of the dependency
// lifecycle: The lifecycle of the dependency
// constructorName: The name of the constructor func
// valueType: The type of the dependency value
// getInstanceFunc: a function that returns the instance
func NewDependency[TType any](name, lifecycle, constructorName string, valueType reflect.Type, getInstanceFunc func(context.Context) (any, error)) (*EctoDependency, error) {
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
		constructor, ok := ectoreflect.GetMethodByName(valueType, constructorName)
		if ok {
			dep.constructor = constructor
		}
	}

	if getInstanceFunc != nil {
		dep.getInstanceFunc = getInstanceFunc
	}

	return dep, nil
}
