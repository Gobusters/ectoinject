package ectoinject

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestDep interface {
	GetString() string
	GetNumber() int
	IncrementCount() int
}

type TestDep1 struct {
	Text  string
	Num   int
	Count int
}

func (t *TestDep1) Constructor() *TestDep1 {
	t.Text = "test"
	t.Num = 0
	t.Count = 0
	return t
}

func (t *TestDep1) GetString() string {
	return t.Text
}

func (t *TestDep1) GetNumber() int {
	return t.Num
}

func (t *TestDep1) IncrementCount() int {
	t.Count++
	return t.Count
}

type testStruct struct {
	Dep TestDep
}

func TestGetSingleton(t *testing.T) {

	config := DIContainerConfig{
		ID:                       "test 1",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[TestDep, TestDep1](container)
	assert.Nil(t, err, "error registering dependency singleton")

	err = RegisterSingleton[testStruct, testStruct](container)
	assert.Nil(t, err, "error registering test struct singleton")

	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	test, err := GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")

	assert.NotNil(t, test.Dep, "Dependency was not set")

	val := test.Dep.GetNumber()
	assert.Equal(t, 0, val)

	val = test.Dep.IncrementCount()
	assert.Equal(t, 1, val)

	// get dep again
	test, err = GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")

	assert.NotNil(t, test.Dep, "Dependency was not set")
	val = test.Dep.IncrementCount()
	assert.Equal(t, 2, val)
}

func TestGetNamedSingleton(t *testing.T) {
	config := DIContainerConfig{
		ID:                       "test 2",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	err = RegisterNamedSingleton[TestDep, TestDep1](container, "foo")
	assert.Nil(t, err, "error registering dependency singleton")

	testDepVal, err := GetDependency[TestDep](ctx, "foo")
	assert.Nil(t, err, "error getting test struct")

	val := testDepVal.IncrementCount()
	assert.Equal(t, 1, val)

	// get dep again
	testDepVal, err = GetDependency[TestDep](ctx, "foo")
	assert.Nil(t, err, "error getting test struct")

	val = testDepVal.IncrementCount()
	assert.Equal(t, 2, val)
}

func TestGetDIContainer(t *testing.T) {
	type testStruct struct {
		Dep DIContainer `inject:""`
	}

	config := DIContainerConfig{
		ID:                       "test 3",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	err = RegisterSingleton[testStruct, testStruct](container)
	assert.Nil(t, err, "error registering test struct singleton")

	test, err := GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")

	assert.NotNil(t, test.Dep, "Dependency was not set")
}

func TestGetScoped(t *testing.T) {
	type testStruct struct {
		Dep TestDep `inject:""`
	}

	config := DIContainerConfig{
		ID:                       "test 4",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	// register deps
	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterScoped[TestDep, TestDep1](container)
	assert.Nil(t, err, "error registering dependency singleton")

	err = RegisterTransient[testStruct, testStruct](container)
	assert.Nil(t, err, "error registering test struct singleton")

	// scope ctx
	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	// get dep
	test, err := GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")
	assert.NotNil(t, test.Dep, "Dependency was not set")

	// increment count
	test.Dep.IncrementCount()
	result := test.Dep.IncrementCount()
	assert.Equal(t, 2, result)

	// get dep again
	test, err = GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")
	assert.NotNil(t, test.Dep, "Dependency was not set")
	result = test.Dep.IncrementCount()
	assert.Equal(t, 3, result)

	// create new scope
	ctx2 := context.Background()
	ctx2, _ = SetActiveContainer(ctx2, config.ID)

	// get dep
	test, err = GetDependency[testStruct](ctx2)
	assert.Nil(t, err, "error getting test struct")
	assert.NotNil(t, test.Dep, "Dependency was not set")

	result = test.Dep.IncrementCount()
	assert.Equal(t, 1, result)
}

func TestGetTransient(t *testing.T) {
	type testStruct struct {
		Dep TestDep `inject:""`
	}

	config := DIContainerConfig{
		ID:                       "test 5",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	// register deps
	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterTransient[TestDep, TestDep1](container)
	assert.Nil(t, err, "error registering dependency singleton")

	err = RegisterTransient[testStruct, testStruct](container)
	assert.Nil(t, err, "error registering test struct singleton")

	// scope ctx
	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)
	ctx = scopeContext(ctx)
	defer unscopeContext(ctx)

	// get dep
	test, err := GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")
	assert.NotNil(t, test.Dep, "Dependency was not set")

	// increment count
	test.Dep.IncrementCount()
	result := test.Dep.IncrementCount()
	assert.Equal(t, 2, result)

	// get dep again
	test, err = GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")
	assert.NotNil(t, test.Dep, "Dependency was not set")
	result = test.Dep.IncrementCount()
	assert.Equal(t, 1, result)
}

func TestUnsafeDependencies(t *testing.T) {
	type testStruct struct {
		dep TestDep `inject:""`
	}

	config := DIContainerConfig{
		ID:                       "test 6",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  true,
	}

	// register deps
	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterTransient[TestDep, TestDep1](container)
	assert.Nil(t, err, "error registering dependency singleton")

	err = RegisterTransient[testStruct, testStruct](container)
	assert.Nil(t, err, "error registering test struct singleton")

	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)
	test, err := GetDependency[testStruct](ctx)
	assert.Nil(t, err, "error getting test struct")

	assert.NotNil(t, test.dep, "Dependency was not set")
}

type bottomStruct struct {
	Dep TestDep `inject:"foo"`
}

type topStruct struct {
	Text  string
	Num   int
	Count int
	Dep   bottomStruct `inject:""`
}

func (t *topStruct) GetString() string {
	return t.Text
}

func (t *topStruct) GetNumber() int {
	return t.Num
}

func (t *topStruct) IncrementCount() int {
	t.Count++
	return t.Count
}

func TestCircularDependency(t *testing.T) {
	config := DIContainerConfig{
		ID:                       "test circular",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterNamedSingleton[TestDep, topStruct](container, "foo")
	assert.Nil(t, err, "error registering topStruct dependency singleton")

	err = RegisterSingleton[bottomStruct, bottomStruct](container)
	assert.Nil(t, err, "error registering bottomStruct dependency singleton")

	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	_, err = GetDependency[TestDep](ctx, "foo")
	assert.NotNil(t, err, "No error getting circular dependency")
	assert.Equal(t, "circular dependency detected for 'foo'. Dependency chain: foo -> github.com/Gobusters/ectoinject.bottomStruct -> foo", err.Error())
}

type testConstructorStruct struct {
	Val string
}

func (t *testConstructorStruct) Constructor(dep TestDep) *testConstructorStruct {
	t.Val = dep.GetString()
	return t
}

func TestConstructor(t *testing.T) {
	config := DIContainerConfig{
		ID:                       "test constructor",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[TestDep, TestDep1](container)
	assert.Nil(t, err, "error registering dependency singleton")

	err = RegisterSingleton[testConstructorStruct, testConstructorStruct](container)
	assert.Nil(t, err, "error registering testConstructorStruct dependency singleton")

	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	testStruct, err := GetDependency[testConstructorStruct](ctx)
	assert.Nil(t, err, "error getting testConstructorStruct")

	assert.Equal(t, "test", testStruct.Val)
}
