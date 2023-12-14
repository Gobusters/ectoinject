package dependency

import (
	"context"
	"reflect"
)

// Dependency is the interface for a dependency. Provides methods for getting and setting the value of the dependency
type Dependency interface {
	SetValue(reflect.Value) error                        // SetValue sets the value of the dependency
	GetValue() reflect.Value                             // GetValue gets the value of the dependency
	GetInstance() (any, error)                           // GetInstance gets the instance of the dependency
	HasValue() bool                                      // HasValue checks if the dependency has a value
	HasConstructor() bool                                // HasConstructor checks if the dependency has a constructor func
	GetConstructor() reflect.Method                      // GetConstructor gets the constructor func of the dependency
	GetInstanceFunc() func(context.Context) (any, error) // GetInstanceFunc returns the custom instance func of the dependency
	GetDependencyType() reflect.Type                     // GetDependencyType returns the type of the dependency
	GetName() string                                     // GetName returns the name of the dependency
	GetLifecycle() string                                // GetLifecycle returns the lifecycle of the dependency
	GetDependencyValueType() reflect.Type                // GetDependencyValueType gets the type of the dependency value
}
