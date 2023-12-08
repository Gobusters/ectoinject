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
		depName = reflectName
	}

	// if the dependency is DIContainer, return the container
	if reflectName == ectoreflect.GetIntefaceName[EctoContainer]() {
		id := defaultContainerID
		if depName != "" {
			id = depName
		}
		depContainer := getContainer(id)
		return ectoreflect.Cast[T](depContainer)
	}

	depValue, err := container.Get(ctx, depName)
	if err != nil {
		return val, err
	}

	val, err = ectoreflect.Cast[T](depValue)
	if err != nil {
		return val, err
	}

	return val, nil
}

func getDependency(ctx context.Context, container *EctoContainer, dep Dependency, chain []Dependency) (Dependency, error) {
	defer func() {
		container.container[dep.GetDependencyName()] = dep // update the container with the new instance
	}()

	// check for circular dependency
	err := checkForCircularDependency(dep.GetDependencyName(), chain)
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
	if dep.HasValue() && dep.GetLifecycle() == lifecycles.Singleton {
		return dep, nil
	}

	// if the user has provided a GetInstanceFunc, use that to get the instance
	instanceFunc := dep.GetInstanceFunc()
	if dep.GetInstanceFunc() != nil {
		instance, err := instanceFunc(ctx, container)
		if err != nil {
			return dep, err
		}

		err = dep.SetValue(instance)
		if err != nil {
			return dep, err
		}
		return dep, nil
	}

	// use the dependency's constructor if it has one
	if dep.HasConstructor() {
		return useDependencyConstructor(ctx, container, dep, chain)
	}

	// if the dependency is a scoped, check the scoped cache
	if dep.GetLifecycle() == lifecycles.Scoped {
		scopedID := getScopedID(ctx)

		// check the scoped cache for an instance
		instance, ok := cache.GetScopedInstance(scopedID, dep.GetDependencyName())
		if !ok {
			// create a new instance
			var err error
			instance, err = getDependencyWithDependencies(ctx, container, dep, chain)
			if err != nil {
				return dep, err
			}

			// add the instance to the scoped cache
			cache.AddScopedInstance(scopedID, dep.GetDependencyName(), instance)
		}

		err = dep.SetValue(instance)
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

func getDependencyWithDependencies(ctx context.Context, container *EctoContainer, dep Dependency, chain []Dependency) (Dependency, error) {
	valueType := dep.GetDependencyValueType()
	// create a new struct value for the dependency
	if valueType.Kind() != reflect.Struct {
		return dep, fmt.Errorf("dependency '%s' has type '%s' which is not a struct", dep.GetDependencyName(), valueType.Name())
	}

	val, err := ectoreflect.NewStructInstance(valueType)
	if err != nil {
		return dep, fmt.Errorf("failed to create new struct instance for dependency '%s': %w", dep.GetDependencyName(), err)
	}
	dep.SetValue(val)

	// Set dependencies
	dep, err = setDependencies(ctx, container, dep, chain)
	if err != nil {
		return dep, err
	}

	return dep, nil
}

func setDependencies(ctx context.Context, container *EctoContainer, dep Dependency, chain []Dependency) (Dependency, error) {
	val := dep.GetValue()

	// check if the dependency is a pointer
	if val.Kind() != reflect.Ptr {
		if !val.CanAddr() {
			return dep, fmt.Errorf("failed to get address of struct instance for dependency '%s'", dep.GetDependencyName())
		}
		// if the dependency is not a pointer, get the pointer to the value
		val = val.Addr()
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return dep, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.GetDependencyName(), val.Kind())
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
		if reflectName == ectoreflect.GetIntefaceName[EctoContainer]() {
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
			msg := fmt.Sprintf("%s has a dependency on %s, but it is not registered", dep.GetDependencyName(), typeName)
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

		ectoreflect.SetField(val, field, childDep.GetValue())
	}

	dep.SetValue(val.Interface())
	return dep, nil
}

func checkForCircularDependency(depName string, chain []Dependency) error {
	for _, dep := range chain {
		if dep.GetDependencyName() == depName {
			depChain := ""
			for _, dep := range chain {
				depChain += fmt.Sprintf("%s -> ", dep.GetDependencyName())
			}
			return fmt.Errorf("circular dependency detected for '%s'. Dependency chain: %s%s", depName, depChain, depName)
		}
	}
	return nil
}

func validateLifecycles(dep Dependency, chain []Dependency) error {
	if dep.GetLifecycle() == lifecycles.Transient {
		// check if any of the parent dependencies are scoped or singleton
		for _, parent := range chain {
			if parent.GetLifecycle() == lifecycles.Scoped || parent.GetLifecycle() == lifecycles.Singleton {
				return fmt.Errorf("captive dependency error: transient dependency %s has %s dependency on %s", dep.GetDependencyName(), parent.GetLifecycle(), parent.GetDependencyName())
			}
		}
	}

	if dep.GetLifecycle() == lifecycles.Scoped {
		// check if any of the parent dependencies are singleton
		for _, parent := range chain {
			if parent.GetLifecycle() == lifecycles.Singleton {
				return fmt.Errorf("captive dependency error: scoped dependency %s has %s dependency on %s", dep.GetDependencyName(), parent.GetLifecycle(), parent.GetDependencyName())
			}
		}
	}

	return nil
}
