package ectoinject

import (
	"fmt"
	"reflect"

	ectoreflect "github.com/Gobusters/ectoinject/internal/reflect"
	"github.com/Gobusters/ectoinject/lifecycles"
)

type DependencyBranch struct {
	Name       string
	Dependency *Dependency
	Tree       DependencyTree
}

type DependencyTree map[string]DependencyBranch

func (t *DependencyTree) ValidateLifecycles(dependency Dependency) error {
	for _, branch := range *t {
		if branch.Dependency == nil {
			continue
		}

		if dependency.lifecycle == lifecycles.Singleton {
			// child must be a singleton
			if branch.Dependency.lifecycle != lifecycles.Singleton {
				return fmt.Errorf("captive dependency error: singleton dependency %s has %s dependency on %s", dependency.dependencyName, branch.Dependency.lifecycle, branch.Dependency.dependencyName)
			}
		}

		if dependency.lifecycle == lifecycles.Scoped {
			// child must be a singleton or scoped
			if branch.Dependency.lifecycle != lifecycles.Singleton && branch.Dependency.lifecycle != lifecycles.Scoped {
				return fmt.Errorf("captive dependency error: scoped dependency %s has %s dependency on %s", dependency.dependencyName, branch.Dependency.lifecycle, branch.Dependency.dependencyName)
			}
		}

		err := branch.Tree.ValidateLifecycles(*branch.Dependency)
		if err != nil {
			return err
		}
	}

	return nil
}

type DependencyChain []string

func (list DependencyChain) Contains(name string) bool {
	for _, item := range list {
		if item == name {
			return true
		}
	}

	return false
}

func getDependencyTreeFromType(container *DIContainer, depName string, t reflect.Type, list DependencyChain) (DependencyTree, error) {
	tree := make(DependencyTree)

	// get depName
	if depName == "" {
		depName = ectoreflect.GetReflectTypeName(t)
	}

	// Check if this dependency is already in the list
	if list.Contains(depName) {
		return tree, fmt.Errorf("circular dependency detected: %s", depName)
	}

	// Add this dependency to the list
	list = append(list, depName)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Process only struct types
	if t.Kind() != reflect.Struct {
		return tree, nil
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("inject")
		fieldType := field.Type

		// Dereference pointer types if needed
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if (tag == "" && container.RequireInjectTag) || tag == "-" {
			continue // skip this field
		}

		typeName := tag
		if typeName == "" {
			typeName = ectoreflect.GetReflectTypeName(fieldType)
		}

		dep := container.container[typeName]

		childTree, err := getDependencyTreeFromType(container, typeName, fieldType, list)
		if err != nil {
			return tree, err
		}

		// Add this dependency to the tree
		branch := DependencyBranch{
			Name: typeName,
			Tree: childTree,
		}
		if dep.dependencyName != "" {
			branch.Dependency = &dep
		}

		tree[typeName] = branch
	}

	return tree, nil
}

func GetDependencyTreeFromType(container *DIContainer, t reflect.Type, name ...string) (DependencyTree, error) {
	depName := ""
	if len(name) > 0 {
		depName = name[0]
	}
	list := DependencyChain{}
	return getDependencyTreeFromType(container, depName, t, list)
}

func GetDependencyTree[T any](container *DIContainer, name ...string) (DependencyTree, error) {
	depName := ""
	if len(name) > 0 {
		depName = name[0]
	}

	t := reflect.TypeOf((*T)(nil)).Elem()
	return GetDependencyTreeFromType(container, t, depName)
}
