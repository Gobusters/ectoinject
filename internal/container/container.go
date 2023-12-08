package container

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Gobusters/ectoinject/container"
	"github.com/Gobusters/ectoinject/dependency"
	"github.com/Gobusters/ectoinject/internal/logging"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/internal/scope"
	"github.com/Gobusters/ectoinject/lifecycles"
)

// Container for dependencies
type EctoContainer struct {
	container.DIContainerConfig                                  // The configuration for the container
	logger                      *logging.Logger                  // The logger to use
	container                   map[string]dependency.Dependency // The container of dependencies
}

func NewEctoContainer(config container.DIContainerConfig, logger *logging.Logger) *EctoContainer {
	return &EctoContainer{
		DIContainerConfig: config,
		logger:            logger,
		container:         make(map[string]dependency.Dependency),
	}
}

func (container *EctoContainer) AddDependency(dep dependency.Dependency) {
	container.container[dep.GetName()] = dep
}

func (container *EctoContainer) GetConstructorFuncName() string {
	return container.ConstructorFuncName
}

func (container *EctoContainer) Get(ctx context.Context, name string) (any, error) {
	ctx = scope.ScopeContext(ctx)   // starts a scope for the dependency tree
	defer scope.UnscopeContext(ctx) // ends the scope for the dependency tree

	// check if the dependency is registered
	dep, ok := container.container[name]
	if !ok {
		return nil, fmt.Errorf("dependency for %s not found", name)
	}

	// get the instance of the dependency
	dep, err := getDependency(ctx, container, dep, []dependency.Dependency{})
	if err != nil {
		return nil, err
	}

	// check if the dependency has a value
	if !dep.HasValue() {
		return nil, fmt.Errorf("dependency for %s is nil", name)
	}

	// return the value
	return dep.GetInstance()
}

func getDependency(ctx context.Context, container *EctoContainer, dep dependency.Dependency, chain []dependency.Dependency) (dependency.Dependency, error) {
	defer func() {
		container.container[dep.GetName()] = dep // update the container with the new instance
	}()

	// check for circular dependency
	err := checkForCircularDependency(dep.GetName(), chain)
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
	if instanceFunc != nil {
		instance, err := instanceFunc(ctx)
		if err != nil {
			return dep, err
		}

		err = dep.SetValue(reflect.ValueOf(instance))
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
		scopedID := scope.GetScopedID(ctx)

		// check the scoped cache for an instance
		var ok bool
		dep, ok = scope.GetScopedDependency(scopedID, dep.GetName())
		if !ok {
			// create a new instance
			var err error
			dep, err = getDependencyWithDependencies(ctx, container, dep, chain)
			if err != nil {
				return dep, err
			}

			// add the instance to the scoped cache
			scope.AddScopedDependency(scopedID, dep)
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

func getDependencyWithDependencies(ctx context.Context, container *EctoContainer, dep dependency.Dependency, chain []dependency.Dependency) (dependency.Dependency, error) {
	valueType := dep.GetDependencyValueType()
	// create a new struct value for the dependency
	if valueType.Kind() != reflect.Struct {
		return dep, fmt.Errorf("dependency '%s' has type '%s' which is not a struct", dep.GetName(), valueType.Name())
	}

	val, err := ectoreflect.NewStructInstance(valueType)
	if err != nil {
		return dep, fmt.Errorf("failed to create new struct instance for dependency '%s': %w", dep.GetName(), err)
	}
	dep.SetValue(val)

	// Set dependencies
	dep, err = setDependencies(ctx, container, dep, chain)
	if err != nil {
		return dep, err
	}

	return dep, nil
}

func setDependencies(ctx context.Context, container *EctoContainer, dep dependency.Dependency, chain []dependency.Dependency) (dependency.Dependency, error) {
	val := dep.GetValue()

	// check if the dependency is a pointer
	if val.Kind() != reflect.Ptr {
		if !val.CanAddr() {
			return dep, fmt.Errorf("failed to get address of struct instance for dependency '%s'", dep.GetName())
		}
		// if the dependency is not a pointer, get the pointer to the value
		val = val.Addr()
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return dep, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.GetName(), val.Kind())
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
		if reflectName == "github.com/Gobusters/ectoinject.DIContainer" {
			id := defaultContainerID
			if tag != "" {
				id = tag
			}
			depContainer := GetContainer(id)
			// setValue(i, isPtr, canSet, val, field, reflect.ValueOf(depContainer))
			err := ectoreflect.SetField(val, field, reflect.ValueOf(depContainer))
			if err != nil {
				return dep, fmt.Errorf("failed to set field '%s' on struct instance for dependency '%s': %w", field.Name, dep.GetName(), err)
			}
			continue
		}

		typeName := tag
		if typeName == "" {
			typeName = reflectName
		}

		childDep, ok := container.container[typeName]
		if !ok {
			msg := fmt.Sprintf("%s has a dependency on %s, but it is not registered", dep.GetName(), typeName)
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

		err = ectoreflect.SetField(val, field, childDep.GetValue())
		if err != nil {
			return dep, fmt.Errorf("failed to set field '%s' on struct instance for dependency '%s': %w", field.Name, dep.GetName(), err)
		}
	}

	dep.SetValue(val)
	return dep, nil
}

func checkForCircularDependency(depName string, chain []dependency.Dependency) error {
	for _, dep := range chain {
		if dep.GetName() == depName {
			depChain := ""
			for _, dep := range chain {
				depChain += fmt.Sprintf("%s -> ", dep.GetName())
			}
			return fmt.Errorf("circular dependency detected for '%s'. Dependency chain: %s%s", depName, depChain, depName)
		}
	}
	return nil
}

func validateLifecycles(dep dependency.Dependency, chain []dependency.Dependency) error {
	if dep.GetLifecycle() == lifecycles.Transient {
		// check if any of the parent dependencies are scoped or singleton
		for _, parent := range chain {
			if parent.GetLifecycle() == lifecycles.Scoped || parent.GetLifecycle() == lifecycles.Singleton {
				return fmt.Errorf("captive dependency error: transient dependency %s has %s dependency on %s", dep.GetName(), parent.GetLifecycle(), parent.GetName())
			}
		}
	}

	if dep.GetLifecycle() == lifecycles.Scoped {
		// check if any of the parent dependencies are singleton
		for _, parent := range chain {
			if parent.GetLifecycle() == lifecycles.Singleton {
				return fmt.Errorf("captive dependency error: scoped dependency %s has %s dependency on %s", dep.GetName(), parent.GetLifecycle(), parent.GetName())
			}
		}
	}

	return nil
}
