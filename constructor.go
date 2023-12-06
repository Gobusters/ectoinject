package ectoinject

import (
	"context"
	"fmt"
	"reflect"

	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
)

func getInstanceFromConstructor(ctx context.Context, container *DIContainer, dep Dependency, chain []Dependency) (any, error) {
	constructor := dep.constructor

	// get the number of args for the constructor
	argCount := constructor.Type.NumIn()
	// create the args for the constructor. The first arg is the struct instance
	args := make([]reflect.Value, argCount)

	// loop through the constructor args and get the instances
	for i := 0; i < argCount; i++ {
		// get the type of the constructor arg. Add 1 to the index to skip the struct instance
		paramType := constructor.Type.In(i)
		// check if the param is a pointer
		isPtr := paramType.Kind() == reflect.Ptr
		if isPtr {
			// if the param is a pointer, get the type of the value
			paramType = paramType.Elem()
		}

		if i == 0 {
			// the first arg is the struct instance
			val, err := ectoreflect.NewStructInstance(paramType)
			if err != nil {
				return nil, err
			}

			if isPtr {
				// if the param is a pointer, get the pointer to the value
				val = val.Addr()
			}
			args[i] = val
			continue
		}
		// get the name of the param type
		paramTypeName := ectoreflect.GetReflectTypeName(paramType)

		// check if the param is a dependency
		childDep, ok := container.container[paramTypeName]
		if !ok {
			return nil, fmt.Errorf("dependency '%s' has unregistered dependency '%s' in '%s' func", dep.dependencyName, paramTypeName, constructor.Name)
		}

		// get the instance of the dependency
		childDep, err := getInstanceOfDependency(ctx, container, childDep, chain)
		if err != nil {
			return nil, err
		}

		// add the dependency to the args
		args[i] = reflect.ValueOf(childDep.instance)
	}

	// call the constructor with the args
	result := constructor.Func.Call(args)

	if len(result) == 0 {
		return nil, fmt.Errorf("constructor '%s' on dependnecy '%s' did not return an instance", constructor.Name, dep.dependencyName)
	}

	instance := result[0].Interface()

	dep.instance = instance
	container.container[dep.dependencyName] = dep

	if len(result) == 1 {
		return instance, nil
	}

	err, ok := result[1].Interface().(error)
	if ok {
		return instance, err
	}

	return instance, nil
}
