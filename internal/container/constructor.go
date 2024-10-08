package container

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Gobusters/ectoinject/dependency"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
)

func useDependencyConstructor(ctx context.Context, container *EctoContainer, dep dependency.Dependency, chain []dependency.Dependency) (context.Context, dependency.Dependency, error) {
	constructor := dep.GetConstructor()

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
				return ctx, dep, err
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

		// if the arg is a context, add the context to the args
		if paramTypeName == "context.Context" {
			args[i] = reflect.ValueOf(ctx)
			continue
		}

		// check if the dependency is the container
		containerDep, ok := container.getContainerDependency(paramTypeName)
		if ok {
			// get the instance of the container
			args[i] = reflect.ValueOf(containerDep)
			continue
		}

		// check if the param is a dependency
		childDep, ok := container.container[paramTypeName]
		if !ok {
			return ctx, dep, fmt.Errorf("dependency '%s' has unregistered dependency '%s' in '%s' func", dep.GetName(), paramTypeName, constructor.Name)
		}

		// get the instance of the dependency
		var err error
		ctx, childDep, err = container.getDependency(ctx, childDep, chain)
		if err != nil {
			return ctx, dep, err
		}

		val := childDep.GetValue()
		// check if the param is a pointer
		if isPtr || paramType.Kind() == reflect.Interface {
			// if the param is a pointer, get the pointer to the value
			val = val.Addr()
		}

		// add the dependency to the args
		args[i] = val
	}

	// call the constructor with the args
	result := constructor.Func.Call(args)

	if len(result) == 0 {
		return ctx, dep, fmt.Errorf("constructor '%s' on dependnecy '%s' did not return an instance", constructor.Name, dep.GetName())
	}

	_ = dep.SetValue(result[0])
	container.container[dep.GetName()] = dep

	if len(result) == 1 {
		return ctx, dep, nil
	}

	err, ok := result[1].Interface().(error)
	if ok {
		return ctx, dep, err
	}

	return ctx, dep, nil
}
