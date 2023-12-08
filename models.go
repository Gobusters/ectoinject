package ectoinject

import (
	"context"
	"reflect"
)

type Dependency interface {
	SetValue(instance any) error
	GetValue() reflect.Value
	HasValue() bool
	HasConstructor() bool
	GetConstructor() reflect.Method
	GetInstanceFunc() InstanceFunc
	GetDependencyType() reflect.Type
	GetDependencyName() string
	GetLifecycle() string
	GetDependencyValueType() reflect.Type
}

type DIContainer interface {
	Get(ctx context.Context, name string) (any, error)
}
