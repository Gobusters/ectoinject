package container

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Gobusters/ectoinject/dependency"
	"github.com/Gobusters/ectoinject/ectocontainer"
	"github.com/Gobusters/ectoinject/internal/logging"
	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/internal/scope"
	"github.com/Gobusters/ectoinject/internal/store"
	"github.com/Gobusters/ectoinject/lifecycles"
)

// Container for dependencies
type EctoContainer struct {
	ectocontainer.DIContainerConfig                                  // The configuration for the container
	logger                          *logging.Logger                  // The logger to use
	container                       map[string]dependency.Dependency // The container of dependencies
}

func NewEctoContainer(config ectocontainer.DIContainerConfig, logger *logging.Logger) *EctoContainer {
	return &EctoContainer{
		DIContainerConfig: config,
		logger:            logger,
		container:         make(map[string]dependency.Dependency),
	}
}

func (container *EctoContainer) GetContainerID() string {
	return container.ID
}

func (container *EctoContainer) AddDependency(dep dependency.Dependency) {
	container.container[dep.GetName()] = dep
}

func (container *EctoContainer) GetConstructorFuncName() string {
	return container.ConstructorFuncName
}

func (container *EctoContainer) Get(ctx context.Context, name string) (context.Context, any, error) {
	if name == "" {
		return ctx, nil, fmt.Errorf("dependency name cannot be empty")
	}

	// check if the dependency is the container
	containerDep, ok := container.getContainerDependency(name)
	if ok {
		return ctx, containerDep, nil
	}

	// check if the dependency is registered
	dep, ok := container.container[name]
	if !ok {
		return ctx, nil, fmt.Errorf("dependency for %s not found", name)
	}

	// get the instance of the dependency
	var err error
	ctx, dep, err = container.getDependency(ctx, dep, []dependency.Dependency{})
	if err != nil {
		return ctx, nil, err
	}

	// check if the dependency has a value
	if !dep.HasValue() {
		return ctx, nil, fmt.Errorf("dependency for %s is nil", name)
	}

	// return the value
	instance, err := dep.GetInstance()
	return ctx, instance, err
}

func (container *EctoContainer) getDependency(ctx context.Context, dep dependency.Dependency, chain []dependency.Dependency) (context.Context, dependency.Dependency, error) {
	defer func() {
		container.container[dep.GetName()] = dep // update the container with the new instance
	}()

	// check for circular dependency
	err := checkForCircularDependency(dep.GetName(), chain)
	if err != nil {
		return ctx, dep, err
	}

	// validate lifecycles
	err = container.validateLifecycles(dep, chain)
	if err != nil {
		// Are captive dependencies allowed?
		if container.AllowCaptiveDependencies {
			container.logger.Warn(err.Error()) // log the error
		} else {
			return ctx, dep, err // return the error
		}
	}

	// add this dependency to the chain
	chain = append(chain, dep)

	// if the dependency is a singleton and dependency has a value already, return the value
	if dep.HasValue() && dep.GetLifecycle() == lifecycles.Singleton {
		return ctx, dep, nil
	}

	// if the dependency is a scoped, check the scoped cache
	if dep.GetLifecycle() == lifecycles.Scoped {
		// check the scoped cache
		scopedDep, ok := scope.GetScopedDependency(ctx, dep.GetName())
		if ok {
			return ctx, scopedDep, nil // return the scoped dependency
		}

		// create a new instance
		ctx, dep, err = container.getDependencyWithDependencies(ctx, dep, chain)
		if err != nil {
			return ctx, dep, err
		}

		// add the instance to the scoped cache
		ctx = scope.AddScopedDependency(ctx, dep)
		return ctx, dep, nil
	}

	// if the user has provided a GetInstanceFunc, use that to get the instance
	instanceFunc := dep.GetInstanceFunc()
	if instanceFunc != nil {
		instance, err := instanceFunc(ctx)
		if err != nil {
			return ctx, dep, err
		}

		err = dep.SetValue(reflect.ValueOf(instance))
		if err != nil {
			return ctx, dep, err
		}
		return ctx, dep, nil
	}

	// use the dependency's constructor if it has one
	if dep.HasConstructor() {
		return useDependencyConstructor(ctx, container, dep, chain)
	} else if container.RequireConstructor {
		container.logger.Warn("dependency '%s' does not have a constructor", dep.GetName())
		return ctx, dep, nil
	}

	// create an instance of the dependency
	ctx, dep, err = container.getDependencyWithDependencies(ctx, dep, chain)
	if err != nil {
		return ctx, dep, err
	}

	return ctx, dep, nil
}

func (container *EctoContainer) getDependencyWithDependencies(ctx context.Context, dep dependency.Dependency, chain []dependency.Dependency) (context.Context, dependency.Dependency, error) {
	valueType := dep.GetDependencyValueType()
	// create a new struct value for the dependency
	if valueType.Kind() != reflect.Struct {
		return ctx, dep, fmt.Errorf("dependency '%s' has type '%s' which is not a struct", dep.GetName(), valueType.Name())
	}

	val, err := ectoreflect.NewStructInstance(valueType)
	if err != nil {
		return ctx, dep, fmt.Errorf("failed to create new struct instance for dependency '%s': %w", dep.GetName(), err)
	}
	dep.SetValue(val)

	// Set dependencies
	ctx, dep, err = container.setDependencies(ctx, dep, chain)
	if err != nil {
		return ctx, dep, err
	}

	return ctx, dep, nil
}

func (container *EctoContainer) setDependencies(ctx context.Context, dep dependency.Dependency, chain []dependency.Dependency) (context.Context, dependency.Dependency, error) {
	val := dep.GetValue()

	// check if the dependency is a pointer
	if val.Kind() != reflect.Ptr {
		if !val.CanAddr() {
			return ctx, dep, fmt.Errorf("failed to get address of struct instance for dependency '%s'", dep.GetName())
		}
		// if the dependency is not a pointer, get the pointer to the value
		val = val.Addr()
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return ctx, dep, fmt.Errorf("instance of dependency '%s' must be a pointer to a struct but is %s", dep.GetName(), val.Kind())
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

		typeName := tag
		if typeName == "" {
			typeName = reflectName
		}

		// check if the dependency is the container
		containerDep, ok := container.getContainerDependency(typeName)
		if ok {
			err := ectoreflect.SetField(val, field, reflect.ValueOf(containerDep))
			if err != nil {
				return ctx, dep, fmt.Errorf("failed to set field '%s' on struct instance for dependency '%s': %w", field.Name, dep.GetName(), err)
			}
			continue
		}

		childDep, ok := container.container[typeName]
		if !ok {
			msg := fmt.Sprintf("%s has a dependency on %s, but it is not registered", dep.GetName(), typeName)
			if container.AllowMissingDependencies {
				container.logger.Info(msg)
				continue
			}
			return ctx, dep, fmt.Errorf(msg)
		}

		var err error
		ctx, childDep, err = container.getDependency(ctx, childDep, chain)
		if err != nil {
			return ctx, dep, err
		}

		container.container[typeName] = childDep

		err = ectoreflect.SetField(val, field, childDep.GetValue())
		if err != nil {
			return ctx, dep, fmt.Errorf("failed to set field '%s' on struct instance for dependency '%s': %w", field.Name, dep.GetName(), err)
		}
	}

	dep.SetValue(val)
	return ctx, dep, nil
}

func (container *EctoContainer) getContainerDependency(name string) (any, bool) {
	if name == ectoreflect.GetIntefaceName[ectocontainer.DIContainer]() {
		return container, true
	}

	dep := store.GetContainer(name)
	if dep != nil {
		return dep, true
	}
	return dep, false
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

func (container *EctoContainer) validateLifecycles(dep dependency.Dependency, chain []dependency.Dependency) error {
	if dep.GetLifecycle() == lifecycles.Transient {
		// check if any of the parent dependencies are scoped or singleton
		for _, parent := range chain {
			if parent.GetLifecycle() == lifecycles.Scoped || parent.GetLifecycle() == lifecycles.Singleton {
				if container.AllowCaptiveDependencies {
					container.logger.Info(fmt.Sprintf("captive dependency: %s is a %s but has a transient dependency %s. %s will behave as a %s", parent.GetName(), parent.GetLifecycle(), dep.GetName(), dep.GetName(), parent.GetLifecycle()))
				} else {
					return fmt.Errorf("captive dependency error: %s is a %s but has a transient dependency %s", parent.GetName(), parent.GetLifecycle(), dep.GetName())
				}
			}
		}
	}

	if dep.GetLifecycle() == lifecycles.Scoped {
		// check if any of the parent dependencies are singleton
		for _, parent := range chain {
			if parent.GetLifecycle() == lifecycles.Singleton {
				if container.AllowCaptiveDependencies {
					container.logger.Info(fmt.Sprintf("captive dependency: %s is a %s but has a scoped dependency %s. %s will behave as a %s", parent.GetName(), parent.GetLifecycle(), dep.GetName(), dep.GetName(), parent.GetLifecycle()))
				} else {
					return fmt.Errorf("captive dependency error: %s is a %s but has a scoped dependency %s", parent.GetName(), parent.GetLifecycle(), dep.GetName())
				}
			}
		}
	}

	return nil
}
