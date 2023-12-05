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
func GetDependency[T any](ctx context.Context, name ...string) (*T, error) {
	container, err := GetActiveContainer(ctx)
	if err != nil {
		return nil, err
	}

	depName := ""
	if len(name) > 0 {
		depName = name[0]
	} else {
		depName = ectoreflect.GetIntefaceName[T]()
	}

	dep, ok := container.container[depName]
	if !ok {
		return nil, fmt.Errorf("dependency for %s not found", depName)
	}

	// get the dependency tree
	tree, err := getDependencyTree[T](*container, depName)
	if err != nil {
		return nil, err
	}

	// validate the dependency tree
	err = tree.ValidateLifecycles(dep)
	if err != nil {
		return nil, err
	}

	dep, err = getInstanceOfDependency(ctx, container, dep)
	if err != nil {
		return nil, err
	}

	if dep.instance == nil {
		return nil, fmt.Errorf("dependency for %s is nil", depName)
	}

	val, ok := dep.instance.(*T)
	if !ok {
		return nil, fmt.Errorf("dependency for %s is not of type %s, actual %T", depName, depName, dep.instance)
	}

	return val, nil
}

func getInstanceOfDependency(ctx context.Context, container *DIContainer, dep Dependency) (Dependency, error) {
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
			instance, err = createInstanceOfDependency(ctx, container, dep.dependencyValueType, dep, "")
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
	instance, err := createInstanceOfDependency(ctx, container, dep.dependencyValueType, dep, "")
	if err != nil {
		return dep, err
	}

	dep.instance = instance

	return dep, nil
}

func createInstanceOfDependency(ctx context.Context, container *DIContainer, t reflect.Type, dep Dependency, parentLifeCycle string) (any, error) {
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
	instanceWithDeps, err := setDependencies(ctx, container, dep, structPtr.Interface())
	if err != nil {
		return nil, err
	}

	return instanceWithDeps, nil
}

func setDependencies(ctx context.Context, container *DIContainer, dep Dependency, v any) (any, error) {
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

		childDep, err := getInstanceOfDependency(ctx, container, childDep)
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

func validateLifecycles(child Dependency, parent Dependency) error {
	if parent.lifecycle == lifecycles.Singleton {
		// child must be a singleton
		if child.lifecycle != lifecycles.Singleton {
			return fmt.Errorf("dependency '%s' is registered as a singleton, but its dependency '%s' is a %s", parent.dependencyName, child.dependencyName, child.lifecycle)
		}
	}

	if parent.lifecycle == lifecycles.Scoped {
		// child must be a singleton or scoped
		if child.lifecycle != lifecycles.Singleton && child.lifecycle != lifecycles.Scoped {
			return fmt.Errorf("dependency '%s' is registered as a scoped, but its dependency '%s' is a %s", parent.dependencyName, child.dependencyName, child.lifecycle)
		}
	}

	return nil
}

var dependencyMap = map[string]DependencyTree{}

func getDependencyTree[T any](container DIContainer, depName string) (DependencyTree, error) {
	// check if the dependency tree is already cached
	if tree, ok := dependencyMap[depName]; ok {
		return tree, nil
	}

	// get the dependency tree
	tree, err := GetDependencyTree[T](&container, depName)
	if err != nil {
		return nil, err
	}

	// cache the dependency tree
	dependencyMap[depName] = tree

	return tree, nil
}
