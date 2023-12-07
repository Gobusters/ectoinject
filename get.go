package ectoinject

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Gobusters/ectoinject/internal/cache"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/lifecycles"
)

// GetDependency gets a dependency from the container
// T: The type of the dependency
// ctx: The context to use. If you're using a scoped dependency, you must use a scoped context from ScopeContext
// name: The name of the dependency. If not provided, the name of the interface is used
func GetDependency[T any](ctx context.Context, name ...string) (T, error) {
	ctx = scopeContext(ctx)   // starts a scope for the dependency tree
	defer unscopeContext(ctx) // ends the scope for the dependency tree
	var val T

	container, err := GetActiveContainer(ctx)
	if err != nil {
		return val, err
	}

	reflectName := ectoreflect.GetIntefaceName[T]()

	depName := ""
	if len(name) > 0 {
		depName = name[0]
	} else {
		// Assuming ectoreflect.GetIntefaceName is a valid function
		depName = reflectName
	}

	// if the dependency is DIContainer, return the container
	if reflectName == ectoreflect.GetIntefaceName[DIContainer]() {
		id := defaultContainerID
		if depName != "" {
			id = depName
		}
		depContainer := getContainer(id)
		return ectoreflect.Cast[T](depContainer)
	}

	// check if the dependency is registered
	dep, ok := container.container[depName]
	if !ok {
		return val, fmt.Errorf("dependency for %s not found", depName)
	}

	// get the instance of the dependency
	dep, err = getDependency(ctx, container, dep, []Dependency{})
	if err != nil {
		return val, err
	}

	// check if the dependency has a value
	if !dep.hasValue() {
		return val, fmt.Errorf("dependency for %s is nil", depName)
	}

	// cast the value to T
	val, err = ectoreflect.CastValue[T](dep.value)
	if err != nil {
		return val, fmt.Errorf("failed to cast dependency for %s: %w", depName, err)
	}

	return val, nil
}

func getDependency(ctx context.Context, container *DIContainer, dep Dependency, chain []Dependency) (Dependency, error) {
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

	// if the dependency is a singleton and dependency has a value already, return the value
	if dep.hasValue() && dep.lifecycle == lifecycles.Singleton {
		return dep, nil
	}

	// if the user has provided a GetInstanceFunc, use that to get the instance
	if dep.getInstanceFunc != nil {
		instance, err := dep.getInstanceFunc(ctx, container)
		if err != nil {
			return dep, err
		}

		err = dep.setInstance(instance)
		if err != nil {
			return dep, err
		}
		return dep, nil
	}

	// use the dependency's constructor if it has one
	if dep.hasConstructor() {
		return useDependencyConstructor(ctx, container, dep, chain)
	}

	// if the dependency is a scoped, check the scoped cache
	if dep.lifecycle == lifecycles.Scoped {
		scopedID := getScopedID(ctx)

		// check the scoped cache for an instance
		instance, ok := cache.GetScopedInstance(scopedID, dep.dependencyName)
		if !ok {
			// create a new instance
			var err error
			instance, err = getDependencyWithDependencies(ctx, container, dep, chain)
			if err != nil {
				return dep, err
			}

			// add the instance to the scoped cache
			cache.AddScopedInstance(scopedID, dep.dependencyName, instance)
		}

		err = dep.setInstance(instance)
		if err != nil {
			return dep, err
		}
		return dep, nil
	}

	// create an instance of the dependency
	dep, err = getDependencyWithDependencies(ctx, container, dep, chain)
	if err != nil {
		return dep, err
	}

	return dep, nil
}

func getDependencyWithDependencies(ctx context.Context, container *DIContainer, dep Dependency, chain []Dependency) (Dependency, error) {
	// create a new struct value for the dependency
	err := dep.createNewStructValue()
	if err != nil {
		return dep, err
	}

	// Set dependencies
	dep, err = setDependencies(ctx, container, dep, chain)
	if err != nil {
		return dep, err
	}

	return dep, nil
}

func setDependencies(ctx context.Context, container *DIContainer, dep Dependency, chain []Dependency) (Dependency, error) {
	val := dep.value

	// check if the dependency is a pointer
	if val.Kind() != reflect.Ptr {
		if !val.CanAddr() {
			return dep, fmt.Errorf("failed to get address of struct instance for dependency '%s'", dep.dependencyName)
		}
		// if the dependency is not a pointer, get the pointer to the value
		val = val.Addr()
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return dep, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.dependencyName, val.Kind())
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(container.InjectTagName)

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
			// setValue(i, isPtr, canSet, val, field, reflect.ValueOf(depContainer))
			ectoreflect.SetField(val, field, reflect.ValueOf(depContainer))
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
			return dep, fmt.Errorf(msg)
		}

		childDep, err := getDependency(ctx, container, childDep, chain)
		if err != nil {
			return dep, err
		}

		container.container[typeName] = childDep

		ectoreflect.SetField(val, field, childDep.value)
	}

	dep.setInstance(val.Interface())
	return dep, nil
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
