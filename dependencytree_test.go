package ectoinject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TreeTestDep1 struct {
	Text *string
}

type TreeTestDep2 struct {
	Num int
}

type TreeTestDep4 struct {
	Dep1 TreeTestDep1 `inject:""`
	Dep2 TreeTestDep2 `inject:""`
}

func TestGetDependencyTree(t *testing.T) {
	type TreeTestStruct struct {
		Dep1 TreeTestDep1  `inject:"foo"`
		Dep2 *TreeTestDep2 `inject:""`
		Dep4 TreeTestDep4  `inject:""`
	}

	config := DIContainerConfig{
		ID:                       "test",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  true,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	RegisterNamedSingleton[TreeTestDep1, TreeTestDep1](container, "foo")

	tree, err := GetDependencyTree[TreeTestStruct](container)
	assert.Nil(t, err, "error getting dependency tree")

	assert.Equal(t, 3, len(tree), "incorrect number of dependencies")

	assert.NotNil(t, tree["foo"].Dependency, "foo dependency not found")
	assert.NotNil(t, tree["github.com/Gobusters/ectoinject.TreeTestDep4"].Tree, "dependency 4 is missing tree")
	assert.NotEmpty(t, tree["github.com/Gobusters/ectoinject.TreeTestDep4"].Tree["github.com/Gobusters/ectoinject.TreeTestDep1"], "dependency 4 is missing dependency 1")
}
