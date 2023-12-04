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
func GetDependency[T any](ctx context.Context) (*T, error) {
	container, err := GetActiveContainer(ctx)
	if err != nil {
		return nil, err
	}

	name := ectoreflect.GetIntefaceName[T]()

	dep, ok := container.container[name]
	if !ok {
		return nil, fmt.Errorf("dependency for %s not found", name)
	}

	dep, err = getDependencyInstance(ctx, container, dep, nil)
	if err != nil {
		return nil, err
	}

	if dep.Instance == nil {
		return nil, fmt.Errorf("dependency for %s is nil", name)
	}

	val, ok := dep.Instance.(*T)
	if !ok {
		return nil, fmt.Errorf("dependency for %s is not of type %s, actual %T", name, name, dep.Instance)
	}

	return val, nil
}

func getDependencyInstance(ctx context.Context, container *DIContainer, dep Dependency, parentDep *Dependency) (Dependency, error) {
	if parentDep != nil {
		// validate lifecycle
		err := validateLifecycles(dep, *parentDep)
		if err != nil {
			if container.AllowCaptiveDependencies {
				container.logger.Warn(err.Error())
			} else {
				return dep, err
			}
		}
	}

	// if the dependency is a singleton and the instance is not nil, return the instance
	if dep.Instance != nil && dep.Lifecycle == lifecycles.Singleton {
		return dep, nil
	}

	// if the user has provided a GetInstanceFunc, use that to get the instance
	if dep.GetInstanceFunc != nil {
		instance, err := dep.GetInstanceFunc(ctx, container)
		if err != nil {
			return dep, err
		}

		dep.Instance = instance
		return dep, nil
	}

	// if the dependency is a scoped, check the scoped cache
	if dep.Lifecycle == lifecycles.Scoped {
		scopedID := getScopedID(ctx)

		// check the scoped cache for an instance
		instance, ok := cache.GetScopedInstance(scopedID, dep.DependencyName)
		if !ok {
			// create a new instance
			var err error
			instance, err = createInstance(ctx, container, dep.DependencyValueType, dep, "")
			if err != nil {
				return dep, err
			}

			// add the instance to the scoped cache
			cache.AddScopedInstance(scopedID, dep.DependencyName, instance)
		}

		dep.Instance = instance
		return dep, nil
	}

	// create an instance of the dependency
	instance, err := createInstance(ctx, container, dep.DependencyValueType, dep, "")
	if err != nil {
		return dep, err
	}

	dep.Instance = instance

	return dep, nil
}

func createInstance(ctx context.Context, container *DIContainer, t reflect.Type, dep Dependency, parentLifeCycle string) (any, error) {
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
		return nil, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.DependencyName, val.Kind())
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.DependencyName, val.Kind())
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
			msg := fmt.Sprintf("%s has a dependency on %s, but it is not registered", dep.DependencyName, typeName)
			if container.AllowMissingDependencies {
				container.logger.Warn(msg)
				continue
			}
			return nil, fmt.Errorf(msg)
		}

		childDep, err := getDependencyInstance(ctx, container, childDep, &dep)
		if err != nil {
			return nil, err
		}

		container.container[typeName] = childDep

		setValue(i, isPtr, canSet, val, field, childDep.Instance)
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
	if parent.Lifecycle == lifecycles.Singleton {
		// child must be a singleton
		if child.Lifecycle != lifecycles.Singleton {
			return fmt.Errorf("dependency '%s' is registered as a singleton, but its dependency '%s' is a %s", parent.DependencyName, child.DependencyName, child.Lifecycle)
		}
	}

	if parent.Lifecycle == lifecycles.Scoped {
		// child must be a singleton or scoped
		if child.Lifecycle != lifecycles.Singleton && child.Lifecycle != lifecycles.Scoped {
			return fmt.Errorf("dependency '%s' is registered as a scoped, but its dependency '%s' is a %s", parent.DependencyName, child.DependencyName, child.Lifecycle)
		}
	}

	return nil
}
