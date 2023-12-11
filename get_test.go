package ectoinject

import (
	"context"
	"testing"

	"github.com/Gobusters/ectoinject/container"
	"github.com/stretchr/testify/assert"
)

type Animal interface {
	Speak() string
}

type Dog struct {
	Name string
}

func (d *Dog) Speak() string {
	return "woof"
}

type Cat struct {
	Name string
}

func (c *Cat) Speak() string {
	return "meow"
}

type Human struct {
	Name  string
	count int
}

func (h *Human) Speak() string {
	return "hello"
}

func (h *Human) Count() int {
	h.count++
	return h.count
}

func TestGetSingleton(t *testing.T) {
	type house struct {
		Dad *Human `inject:""`
		Mom Human  `inject:"mom"`
		Pet Animal `inject:""`
	}

	config := container.DIContainerConfig{
		ID:                       "test get singleton",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Human, Human](container)
	assert.Nil(t, err, "error registering dad depenedency")

	err = RegisterSingleton[Human, Human](container, "mom")
	assert.Nil(t, err, "error registering mom depenedency")

	err = RegisterSingleton[Animal, Dog](container)
	assert.Nil(t, err, "error registering dog depenedency")

	err = RegisterSingleton[*house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	houseInstance, err := GetDependency[*house](ctx)
	assert.Nil(t, err, "error getting house instance")
	assert.NotNil(t, houseInstance.Dad, "dad dependency was not set")

	assert.Equal(t, 1, houseInstance.Dad.Count())
	assert.Equal(t, 1, houseInstance.Mom.Count())

	houseInstance, err = GetDependency[*house](ctx)
	assert.Nil(t, err, "error getting house instance")
	assert.NotNil(t, houseInstance.Dad, "dad dependency was not set")

	assert.Equal(t, 2, houseInstance.Dad.Count())
	assert.Equal(t, 2, houseInstance.Mom.Count())
	assert.Equal(t, "woof", houseInstance.Pet.Speak())
}

func TestGetNamedSingleton(t *testing.T) {
	config := container.DIContainerConfig{
		ID:                       "test get named singleton",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Animal, Human](container, "dad")
	assert.Nil(t, err, "error registering dad depenedency")

	err = RegisterSingleton[Animal, Dog](container, "dog")
	assert.Nil(t, err, "error registering dog depenedency")

	err = RegisterSingleton[Animal, Cat](container, "cat")
	assert.Nil(t, err, "error registering cat depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	dad, err := GetNamedDependency[Animal](ctx, "dad")
	assert.Nil(t, err, "error getting dad instance")
	assert.NotNil(t, dad, "dad dependency was not set")
	assert.Equal(t, "hello", dad.Speak())

	dog, err := GetNamedDependency[Animal](ctx, "dog")
	assert.Nil(t, err, "error getting dog instance")
	assert.NotNil(t, dog, "dog dependency was not set")
	assert.Equal(t, "woof", dog.Speak())

	cat, err := GetNamedDependency[Animal](ctx, "cat")
	assert.Nil(t, err, "error getting cat instance")
	assert.NotNil(t, cat, "cat dependency was not set")
	assert.Equal(t, "meow", cat.Speak())
}

func TestGetDIContainer(t *testing.T) {
	type fakeService struct {
		Dep DIContainer `inject:"test get di container"`
	}

	config := container.DIContainerConfig{
		ID:                       "test get di container",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[fakeService, fakeService](container)
	assert.Nil(t, err, "error registering fake service depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	fakeServiceVal, err := GetDependency[fakeService](ctx)
	assert.Nil(t, err, "error getting fake service")

	assert.NotNil(t, fakeServiceVal.Dep, "fake service dependency was not set")

	containerVal, err := GetNamedDependency[DIContainer](ctx, "test get di container")
	assert.Nil(t, err, "error getting container")

	assert.NotNil(t, containerVal, "container dependency was not found")
	assert.Equal(t, container, containerVal, "container values do not match")
}

// func TestGetScoped(t *testing.T) {
// 	type testStruct struct {
// 		Dep TestDep `inject:""`
// 	}

// 	type testStruct2 struct {
// 		Dep  testStruct `inject:""`
// 		Dep1 TestDep    `inject:""`
// 	}

// 	config := container.DIContainerConfig{
// 		ID:                       "test 4",
// 		AllowCaptiveDependencies: true,
// 		AllowMissingDependencies: true,
// 		RequireInjectTag:         false,
// 		AllowUnsafeDependencies:  false,
// 	}

// 	// register deps
// 	container, err := NewDIContainer(config)
// 	assert.Nil(t, err, "error creating container")

// 	err = RegisterScoped[TestDep, TestDep1](container)
// 	assert.Nil(t, err, "error registering dependency")

// 	err = RegisterTransient[testStruct, testStruct](container)
// 	assert.Nil(t, err, "error registering test struct")

// 	err = RegisterScoped[testStruct2, testStruct2](container)
// 	assert.Nil(t, err, "error registering test struct 2")

// 	// scope ctx
// 	ctx := context.Background()
// 	ctx, _ = SetActiveContainer(ctx, config.ID)

// 	// get dep
// 	testVal, err := GetDependency[testStruct2](ctx)
// 	assert.Nil(t, err, "error getting test struct")
// 	assert.NotNil(t, testVal.Dep, "Test struct dependnecy was not set")
// 	assert.NotNil(t, testVal.Dep1, "Test struct dependnecy was not set")

// 	// increment count
// 	test.Dep.IncrementCount()
// 	result := test.Dep.IncrementCount()
// 	assert.Equal(t, 2, result)

// 	// get dep again
// 	test, err = GetDependency[testStruct](ctx)
// 	assert.Nil(t, err, "error getting test struct")
// 	assert.NotNil(t, test.Dep, "Dependency was not set")
// 	result = test.Dep.IncrementCount()
// 	assert.Equal(t, 3, result)

// 	// create new scope
// 	ctx2 := context.Background()
// 	ctx2, _ = SetActiveContainer(ctx2, config.ID)

// 	// get dep
// 	test, err = GetDependency[testStruct](ctx2)
// 	assert.Nil(t, err, "error getting test struct")
// 	assert.NotNil(t, test.Dep, "Dependency was not set")

// 	result = test.Dep.IncrementCount()
// 	assert.Equal(t, 1, result)
// }

// func TestGetTransient(t *testing.T) {
// 	type testStruct struct {
// 		Dep TestDep `inject:""`
// 	}

// 	config := container.DIContainerConfig{
// 		ID:                       "test 5",
// 		AllowCaptiveDependencies: true,
// 		AllowMissingDependencies: true,
// 		RequireInjectTag:         false,
// 		AllowUnsafeDependencies:  false,
// 	}

// 	// register deps
// 	container, err := NewDIContainer(config)
// 	assert.Nil(t, err, "error creating container")

// 	err = RegisterTransient[TestDep, TestDep1](container)
// 	assert.Nil(t, err, "error registering dependency singleton")

// 	err = RegisterTransient[testStruct, testStruct](container)
// 	assert.Nil(t, err, "error registering test struct singleton")

// 	// scope ctx
// 	ctx := context.Background()
// 	ctx, _ = SetActiveContainer(ctx, config.ID)

// 	// get dep
// 	test, err := GetDependency[testStruct](ctx)
// 	assert.Nil(t, err, "error getting test struct")
// 	assert.NotNil(t, test.Dep, "Dependency was not set")

// 	// increment count
// 	test.Dep.IncrementCount()
// 	result := test.Dep.IncrementCount()
// 	assert.Equal(t, 2, result)

// 	// get dep again
// 	test, err = GetDependency[testStruct](ctx)
// 	assert.Nil(t, err, "error getting test struct")
// 	assert.NotNil(t, test.Dep, "Dependency was not set")
// 	result = test.Dep.IncrementCount()
// 	assert.Equal(t, 1, result)
// }

// func TestUnsafeDependencies(t *testing.T) {
// 	type testStruct struct {
// 		dep TestDep `inject:""`
// 	}

// 	config := container.DIContainerConfig{
// 		ID:                       "test 6",
// 		AllowCaptiveDependencies: true,
// 		AllowMissingDependencies: true,
// 		RequireInjectTag:         false,
// 		AllowUnsafeDependencies:  true,
// 	}

// 	// register deps
// 	container, err := NewDIContainer(config)
// 	assert.Nil(t, err, "error creating container")

// 	err = RegisterTransient[TestDep, TestDep1](container)
// 	assert.Nil(t, err, "error registering dependency singleton")

// 	err = RegisterTransient[testStruct, testStruct](container)
// 	assert.Nil(t, err, "error registering test struct singleton")

// 	ctx := context.Background()
// 	ctx, _ = SetActiveContainer(ctx, config.ID)
// 	test, err := GetDependency[testStruct](ctx)
// 	assert.Nil(t, err, "error getting test struct")

// 	assert.NotNil(t, test.dep, "Dependency was not set")
// }

// type bottomStruct struct {
// 	Dep TestDep `inject:"foo"`
// }

// type topStruct struct {
// 	Text  string
// 	Num   int
// 	Count int
// 	Dep   bottomStruct `inject:""`
// }

// func (t *topStruct) GetString() string {
// 	return t.Text
// }

// func (t *topStruct) GetNumber() int {
// 	return t.Num
// }

// func (t *topStruct) IncrementCount() int {
// 	t.Count++
// 	return t.Count
// }

// func TestCircularDependency(t *testing.T) {
// 	config := container.DIContainerConfig{
// 		ID:                       "test circular",
// 		AllowCaptiveDependencies: true,
// 		AllowMissingDependencies: true,
// 		RequireInjectTag:         false,
// 		AllowUnsafeDependencies:  false,
// 	}

// 	container, err := NewDIContainer(config)
// 	assert.Nil(t, err, "error creating container")

// 	err = RegisterSingleton[TestDep, topStruct](container, "foo")
// 	assert.Nil(t, err, "error registering topStruct dependency singleton")

// 	err = RegisterSingleton[bottomStruct, bottomStruct](container)
// 	assert.Nil(t, err, "error registering bottomStruct dependency singleton")

// 	ctx := context.Background()
// 	ctx, _ = SetActiveContainer(ctx, config.ID)

// 	_, err = GetNamedDependency[TestDep](ctx, "foo")
// 	assert.NotNil(t, err, "No error getting circular dependency")
// 	assert.Equal(t, "circular dependency detected for 'foo'. Dependency chain: foo -> github.com/Gobusters/ectoinject.bottomStruct -> foo", err.Error())
// }

// type testConstructorStruct struct {
// 	Val string
// }

// func (t *testConstructorStruct) Constructor(dep TestDep) *testConstructorStruct {
// 	t.Val = dep.GetString()
// 	return t
// }

// func TestConstructor(t *testing.T) {
// 	config := container.DIContainerConfig{
// 		ID:                       "test constructor",
// 		AllowCaptiveDependencies: true,
// 		AllowMissingDependencies: true,
// 		RequireInjectTag:         false,
// 		AllowUnsafeDependencies:  false,
// 	}

// 	container, err := NewDIContainer(config)
// 	assert.Nil(t, err, "error creating container")

// 	err = RegisterSingleton[TestDep, TestDep1](container)
// 	assert.Nil(t, err, "error registering dependency singleton")

// 	err = RegisterSingleton[testConstructorStruct, testConstructorStruct](container)
// 	assert.Nil(t, err, "error registering testConstructorStruct dependency singleton")

// 	ctx := context.Background()
// 	ctx, _ = SetActiveContainer(ctx, config.ID)

// 	testStruct, err := GetDependency[testConstructorStruct](ctx)
// 	assert.Nil(t, err, "error getting testConstructorStruct")

// 	assert.Equal(t, "test", testStruct.Val)
// }
