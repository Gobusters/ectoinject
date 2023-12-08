package dependency

import (
	"context"
	"reflect"
)

type Dependency interface {
	SetValue(reflect.Value) error
	GetValue() reflect.Value
	GetInstance() (any, error)
	HasValue() bool
	HasConstructor() bool
	GetConstructor() reflect.Method
	GetInstanceFunc() func(context.Context) (any, error)
	GetDependencyType() reflect.Type
	GetName() string
	GetLifecycle() string
	GetDependencyValueType() reflect.Type
}
