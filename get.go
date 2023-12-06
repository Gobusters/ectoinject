package ectoinject

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/Gobusters/ectoinject/internal/cache"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/lifecycles"
)

// GetDependency gets a dependency from the container
// T: The type of the dependency
// ctx: The context to use. If you're using a scoped dependency, you must use a scoped context from ScopeContext
// name: The name of the dependency. If not provided, the name of the interface is used
func GetDependency[T any](ctx context.Context, name ...string) (T, error) {
	var val T
	typeOfT := reflect.TypeOf((*T)(nil)).Elem() // Get the type of T safely

	container, err := GetActiveContainer(ctx)
	if err != nil {
		return val, err
	}

	depName := ""
	if len(name) > 0 {
		depName = name[0]
	} else {
		// Assuming ectoreflect.GetIntefaceName is a valid function
		depName = ectoreflect.GetIntefaceName[T]()
	}

	dep, ok := container.container[depName]
	if !ok {
		return val, fmt.Errorf("dependency for %s not found", depName)
	}

	dep, err = getInstanceOfDependency(ctx, container, dep, []Dependency{})
	if err != nil {
		return val, err
	}

	if dep.instance == nil {
		return val, fmt.Errorf("dependency for %s is nil", depName)
	}

	kind := typeOfT.Kind() // Use the safe type obtained earlier

	// Check for interface or pointer kind
	if kind == reflect.Interface || kind == reflect.Ptr {
		val, ok = dep.instance.(T)
		if !ok {
			return val, fmt.Errorf("dependency for %s is not of type %T, actual %T", depName, val, dep.instance)
		}
		return val, nil
	}

	// Assuming ectoreflect.DereferencePointer is a valid function
	instance := ectoreflect.DereferencePointer(dep.instance)
	val, ok = instance.(T)
	if !ok {
		return val, fmt.Errorf("dependency for %s is not of type %T, actual %T", depName, val, instance)
	}

	return val, nil
}

func getInstanceOfDependency(ctx context.Context, container *DIContainer, dep Dependency, chain []Dependency) (Dependency, error) {
	defer func() {
		container.container[dep.dependencyName] = dep // update the container with the new instance
	}()

	// check for circular dependency
	err := checkForCircularDependency(dep.dependencyName, chain)
	if err != nil {
		return dep, err
	}

	// validate lifecycles
	err = validateLifecycles(dep, chain)
	if err != nil {
		return dep, err
	}

	// add this dependency to the chain
	chain = append(chain, dep)

	// if the dependency is a singleton and the instance is not nil, return the instance
	if dep.instance != nil && dep.lifecycle == lifecycles.Singleton {
		return dep, nil
	}

	// if the user has provided a GetInstanceFunc, use that to get the instance
	if dep.getInstanceFunc != nil {
		instance, err := dep.getInstanceFunc(ctx, container)
		if err != nil {
			return dep, err
		}

		dep.instance = instance
		return dep, nil
	}

	// if the dependency is a scoped, check the scoped cache
	if dep.lifecycle == lifecycles.Scoped {
		scopedID := getScopedID(ctx)

		// check the scoped cache for an instance
		instance, ok := cache.GetScopedInstance(scopedID, dep.dependencyName)
		if !ok {
			// create a new instance
			var err error
			instance, err = createInstanceOfDependency(ctx, container, dep.dependencyValueType, dep, chain)
			if err != nil {
				return dep, err
			}

			// add the instance to the scoped cache
			cache.AddScopedInstance(scopedID, dep.dependencyName, instance)
		}

		dep.instance = instance
		return dep, nil
	}

	// create an instance of the dependency
	instance, err := createInstanceOfDependency(ctx, container, dep.dependencyValueType, dep, chain)
	if err != nil {
		return dep, err
	}

	dep.instance = instance

	return dep, nil
}

func createInstanceOfDependency(ctx context.Context, container *DIContainer, t reflect.Type, dep Dependency, chain []Dependency) (any, error) {
	// use the dependency's constructor if it has one
	if dep.hasConstructor() {
		return getInstanceFromConstructor(ctx, container, dep, chain)
	}

	// Create a new instance of the struct
	val, err := ectoreflect.NewStructInstance(t)
	if err != nil {
		return nil, err
	}

	valInterface := val.Interface()

	// Ensure the value is a struct
	if reflect.ValueOf(valInterface).Kind() != reflect.Struct {
		return nil, fmt.Errorf("createInstance: casted value must be a struct, got %T", valInterface)
	}

	// Take the address of the struct to pass a pointer to setDependencies
	structPtr := reflect.New(reflect.TypeOf(valInterface))
	structPtr.Elem().Set(reflect.ValueOf(valInterface))

	// Set dependencies
	instanceWithDeps, err := setDependencies(ctx, container, dep, structPtr.Interface(), chain)
	if err != nil {
		return nil, err
	}

	return instanceWithDeps, nil
}

func setDependencies(ctx context.Context, container *DIContainer, dep Dependency, v any, chain []Dependency) (any, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return nil, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.dependencyName, val.Kind())
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.dependencyName, val.Kind())
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("inject")
		isPtr := field.Type.Kind() == reflect.Ptr
		if isPtr {
			field.Type = field.Type.Elem()
		}

		if (tag == "" && container.RequireInjectTag) || tag == "-" {
			continue // skip this field
		}

		canSet := val.Field(i).CanSet()

		if !canSet && !container.AllowUnsafeDependencies {
			continue // skip this field
		}

		reflectName := ectoreflect.GetReflectTypeName(field.Type)

		// check if dep is DIContainer
		if reflectName == ectoreflect.GetIntefaceName[DIContainer]() {
			id := defaultContainerID
			if tag != "" {
				id = tag
			}
			depContainer := getContainer(id)
			setValue(i, isPtr, canSet, val, field, *depContainer)
			continue
		}

		typeName := tag
		if typeName == "" {
			typeName = reflectName
		}

		childDep, ok := container.container[typeName]
		if !ok {
			msg := fmt.Sprintf("%s has a dependency on %s, but it is not registered", dep.dependencyName, typeName)
			if container.AllowMissingDependencies {
				container.logger.Warn(msg)
				continue
			}
			return nil, fmt.Errorf(msg)
		}

		childDep, err := getInstanceOfDependency(ctx, container, childDep, chain)
		if err != nil {
			return nil, err
		}

		container.container[typeName] = childDep

		setValue(i, isPtr, canSet, val, field, childDep.instance)
	}

	return v, nil
}

func setValue(index int, isPtr, canSet bool, val reflect.Value, field reflect.StructField, value any) {
	fieldVal := val.Field(index)

	// Convert `value` to a reflect.Value
	reflectValue := reflect.ValueOf(value)

	if isPtr {
		// Create a new pointer to the value
		ptr := reflect.New(reflect.TypeOf(value))
		ptr.Elem().Set(reflectValue)
		reflectValue = ptr

		if reflectValue.Kind() == reflect.Ptr {
			if reflectValue.IsNil() {
				// ignore nil pointers
				return
			}
			// Use the value the pointer points to
			reflectValue = reflectValue.Elem()
		}
	}

	if canSet {
		// Set the value directly if it's settable
		fieldVal.Set(reflectValue)
	} else {
		// If not settable, use unsafe to set the value
		ptr := reflect.NewAt(field.Type, unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		ptr.Set(reflectValue)
	}
}

func checkForCircularDependency(depName string, chain []Dependency) error {
	for _, dep := range chain {
		if dep.dependencyName == depName {
			depChain := ""
			for _, dep := range chain {
				depChain += fmt.Sprintf("%s -> ", dep.dependencyName)
			}
			return fmt.Errorf("circular dependency detected for '%s'. Dependency chain: %s%s", depName, depChain, depName)
		}
	}
	return nil
}

func validateLifecycles(dep Dependency, chain []Dependency) error {
	if dep.lifecycle == lifecycles.Transient {
		// check if any of the parent dependencies are scoped or singleton
		for _, parent := range chain {
			if parent.lifecycle == lifecycles.Scoped || parent.lifecycle == lifecycles.Singleton {
				return fmt.Errorf("captive dependency error: transient dependency %s has %s dependency on %s", dep.dependencyName, parent.lifecycle, parent.dependencyName)
			}
		}
	}

	if dep.lifecycle == lifecycles.Scoped {
		// check if any of the parent dependencies are singleton
		for _, parent := range chain {
			if parent.lifecycle == lifecycles.Singleton {
				return fmt.Errorf("captive dependency error: scoped dependency %s has %s dependency on %s", dep.dependencyName, parent.lifecycle, parent.dependencyName)
			}
		}
	}

	return nil
}
