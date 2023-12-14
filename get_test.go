package ectoinject

import (
	"context"
	"fmt"
	"testing"

	"github.com/Gobusters/ectoinject/ectocontainer"
	"github.com/Gobusters/ectoinject/lifecycles"
	"github.com/stretchr/testify/assert"
)

type Animal interface {
	Speak() string
}

type Dog struct {
}

func (d *Dog) Speak() string {
	return "woof"
}

type Cat struct {
}

func (c *Cat) Speak() string {
	return "meow"
}

type Human struct {
	Name  string
	count int
}

type Person interface {
	Speak() string
	Count() int
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
		Dad Person `inject:""`
		Mom Human  `inject:"mom"`
		Pet Animal `inject:""`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test get singleton",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Person, Human](container)
	assert.Nil(t, err, "error registering dad depenedency")

	err = RegisterSingleton[Human, Human](container, "mom")
	assert.Nil(t, err, "error registering mom depenedency")

	err = RegisterSingleton[Animal, Dog](container)
	assert.Nil(t, err, "error registering dog depenedency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseInstance, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting house instance")
	assert.NotNil(t, houseInstance.Dad, "dad dependency was not set")

	assert.Equal(t, 1, houseInstance.Dad.Count())
	assert.Equal(t, 1, houseInstance.Mom.Count())

	_, houseInstance, err = GetContext[house](ctx)
	assert.Nil(t, err, "error getting house instance")
	assert.NotNil(t, houseInstance.Dad, "dad dependency was not set")

	assert.Equal(t, 2, houseInstance.Dad.Count())
	assert.Equal(t, 1, houseInstance.Mom.Count())
	assert.Equal(t, "woof", houseInstance.Pet.Speak())
}

func TestGetNamedSingleton(t *testing.T) {
	config := ectocontainer.DIContainerConfig{
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

	_, dad, err := GetNamedDependency[Animal](ctx, "dad")
	assert.Nil(t, err, "error getting dad instance")
	assert.NotNil(t, dad, "dad dependency was not set")
	assert.Equal(t, "hello", dad.Speak())

	_, dog, err := GetNamedDependency[Animal](ctx, "dog")
	assert.Nil(t, err, "error getting dog instance")
	assert.NotNil(t, dog, "dog dependency was not set")
	assert.Equal(t, "woof", dog.Speak())

	_, cat, err := GetNamedDependency[Animal](ctx, "cat")
	assert.Nil(t, err, "error getting cat instance")
	assert.NotNil(t, cat, "cat dependency was not set")
	assert.Equal(t, "meow", cat.Speak())
}

func TestGetDIContainer(t *testing.T) {
	type fakeService struct {
		Dep ectocontainer.DIContainer `inject:"test get di container"`
	}

	config := ectocontainer.DIContainerConfig{
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

	_, fakeServiceVal, err := GetContext[fakeService](ctx)
	assert.Nil(t, err, "error getting fake service")

	assert.NotNil(t, fakeServiceVal.Dep, "fake service dependency was not set")

	_, containerVal, err := GetNamedDependency[ectocontainer.DIContainer](ctx, "test get di container")
	assert.Nil(t, err, "error getting container")

	assert.NotNil(t, containerVal, "container dependency was not found")
	assert.Equal(t, container, containerVal, "container values do not match")
}

func TestGetScoped(t *testing.T) {
	type house struct {
		Dad Person `inject:""`
	}

	type city struct {
		House house  `inject:""`
		Mayor Person `inject:""`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test get scoped",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	// register deps
	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterScoped[Person, Human](container)
	assert.Nil(t, err, "error registering dad depenedency")

	err = RegisterTransient[house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	err = RegisterTransient[city, city](container)
	assert.Nil(t, err, "error registering city depenedency")

	// scope ctx
	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	// get dep
	ctx, cityVal, err := GetContext[city](ctx)
	assert.Nil(t, err, "error getting city")
	assert.NotNil(t, cityVal.House, "house dependency was not set")
	assert.NotNil(t, cityVal.Mayor, "mayor dependency was not set")

	// increment count
	assert.Equal(t, 1, cityVal.Mayor.Count())
	assert.Equal(t, 2, cityVal.House.Dad.Count())

	// get dep
	_, cityVal, err = GetContext[city](ctx)
	assert.Nil(t, err, "error getting city")
	assert.NotNil(t, cityVal.House, "house dependency was not set")
	assert.NotNil(t, cityVal.Mayor, "mayor dependency was not set")

	// increment count
	assert.Equal(t, 3, cityVal.Mayor.Count())
	assert.Equal(t, 4, cityVal.House.Dad.Count())

	// create new scope
	ctx = context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	// get dep again
	_, cityVal, err = GetContext[city](ctx)
	assert.Nil(t, err, "error getting city")
	assert.NotNil(t, cityVal.House, "house dependency was not set")
	assert.NotNil(t, cityVal.Mayor, "mayor dependency was not set")

	// increment count again
	assert.Equal(t, 1, cityVal.Mayor.Count())
	assert.Equal(t, 2, cityVal.House.Dad.Count())
}

func TestGetTransient(t *testing.T) {
	type house struct {
		Dad *Human `inject:""`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test get transient",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	// register deps
	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterTransient[Human, Human](container)
	assert.Nil(t, err, "error registering dad dependency")

	err = RegisterTransient[house, house](container)
	assert.Nil(t, err, "error registering house dependency")

	// scope ctx
	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	// get dep
	_, houseVal, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting house struct")
	assert.NotNil(t, houseVal.Dad, "dad dependency was not set")

	// increment count
	assert.Equal(t, 1, houseVal.Dad.Count())

	// get dep again
	_, houseVal, err = GetContext[house](ctx)
	assert.Nil(t, err, "error getting house struct")
	assert.NotNil(t, houseVal.Dad, "dad dependency was not set")

	// increment count again
	assert.Equal(t, 1, houseVal.Dad.Count())
}

type circularAnimal struct {
	Dep Animal `inject:"foo"`
}

func (c *circularAnimal) Speak() string {
	return "circles"
}

func TestCircularDependency(t *testing.T) {
	type house struct {
		Pet Animal `inject:"foo"`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test circular dependency",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Animal, circularAnimal](container, "foo")
	assert.Nil(t, err, "error registering animal dependency singleton")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house dependency singleton")

	ctx := context.Background()
	ctx, _ = SetActiveContainer(ctx, config.ID)

	_, _, err = GetContext[house](ctx)
	assert.NotNil(t, err, "No error getting circular dependency")
	assert.Equal(t, "circular dependency detected for 'foo'. Dependency chain: github.com/Gobusters/ectoinject.house -> foo -> foo", err.Error())
}

type monkey struct {
	repeat string
}

func (m *monkey) Speak() string {
	return m.repeat
}

func (m *monkey) Constructor(dep Animal) *monkey {
	m.repeat = dep.Speak()
	return m
}

func (m *monkey) DIConstructor(ctx context.Context, di ectocontainer.DIContainer) (*monkey, error) {
	_, animal, err := GetDependency[Animal](ctx, di, "github.com/Gobusters/ectoinject.Animal")
	if err != nil {
		return nil, fmt.Errorf("monkey DIConstructor failed to get dependency: %w", err)
	}

	m.repeat = animal.Speak()

	return m, nil
}

func (m *monkey) ErrorConstructor() (*monkey, error) {
	return m, fmt.Errorf("monkey error constructor has returned an error")
}

func TestConstructor(t *testing.T) {
	config := ectocontainer.DIContainerConfig{
		ID:                       "test constructor",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[monkey, monkey](container)
	assert.Nil(t, err, "error registering monkey dependency singleton")

	err = RegisterSingleton[Animal, Dog](container)
	assert.Nil(t, err, "error registering animal dependency singleton")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, monkeyVal, err := GetContext[monkey](ctx)
	assert.Nil(t, err, "error getting monkey struct")

	assert.Equal(t, "woof", monkeyVal.Speak())
}

func TestConstructorWithDIContainer(t *testing.T) {
	config := ectocontainer.DIContainerConfig{
		ID:                       "test constructor with DIContainer",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
		ConstructorFuncName:      "DIConstructor",
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[monkey, monkey](container)
	assert.Nil(t, err, "error registering monkey dependency singleton")

	err = RegisterSingleton[Animal, Cat](container)
	assert.Nil(t, err, "error registering animal dependency singleton")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, monkeyVal, err := GetContext[monkey](ctx)
	assert.Nil(t, err, "error getting monkey struct")

	assert.Equal(t, "meow", monkeyVal.Speak())
}

func TestConstructorWithError(t *testing.T) {
	config := ectocontainer.DIContainerConfig{
		ID:                       "test constructor with error",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
		ConstructorFuncName:      "ErrorConstructor",
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[monkey, monkey](container)
	assert.Nil(t, err, "error registering monkey dependency singleton")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, _, err = GetContext[monkey](ctx)
	assert.Error(t, err, "monkey error constructor has returned an error")
}

func TestGetInstanceFunc(t *testing.T) {
	type house struct {
		Dad Person `inject:""`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test get instance func",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterInstanceFunc[Person](container, lifecycles.Singleton, func(ctx context.Context) (any, error) {
		return &Human{
			Name: "Dave",
		}, nil
	})
	assert.Nil(t, err, "error registering dad depenedency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house dependency singleton")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseVal, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting house struct")

	assert.Equal(t, "Dave", houseVal.Dad.(*Human).Name)
}

func TestInstanceDependency(t *testing.T) {
	type house struct {
		Location string `inject:"foo"`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test instance dependency",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterInstance[string](container, "bar", "foo")
	assert.Nil(t, err, "error registering dad depenedency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house dependency singleton")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseVal, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting house struct")

	assert.Equal(t, "bar", houseVal.Location)
}

func TestRequireInjectTag(t *testing.T) {
	type house struct {
		Dad Person
		Mom Person `inject:"mom"`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test require inject tag",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         true,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Person, Human](container)
	assert.Nil(t, err, "error registering dad depenedency")

	err = RegisterSingleton[Human, Human](container, "mom")
	assert.Nil(t, err, "error registering mom depenedency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseInstance, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting house instance")
	assert.Nil(t, houseInstance.Dad, "dad dependency was set")
	assert.NotNil(t, houseInstance.Mom, "mom dependency was not set")
}

func TestDisableMissingDependencies(t *testing.T) {
	type house struct {
		Dog Animal
		Cat Animal `inject:"fluffy"`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test disable missing dependencies",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: false,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Animal, Dog](container)
	assert.Nil(t, err, "error registering dog depenedency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, _, err = GetContext[house](ctx)
	assert.NotNil(t, err, "no error returned for missing dependency")
	assert.Equal(t, "github.com/Gobusters/ectoinject.house has a dependency on fluffy, but it is not registered", err.Error())
}

func TestEnableMissingDependencies(t *testing.T) {
	type house struct {
		Dog Animal
		Cat Animal `inject:"fluffy"`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test enable missing dependencies",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Animal, Dog](container)
	assert.Nil(t, err, "error registering dog depenedency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseVal, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting house instance")
	assert.Nil(t, houseVal.Cat, "cat dependency was set")
	assert.NotNil(t, houseVal.Dog, "dog dependency was not set")
}

func TestDisableCaptiveDependencies(t *testing.T) {
	type lab struct {
		Dog
		Owner Person `inject:"john"`
	}

	type house struct {
		Dog Animal `inject:"buddy"`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test disable captive dependencies",
		AllowCaptiveDependencies: false,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterTransient[Person, Human](container, "john")
	assert.Nil(t, err, "error registering john depenedency")

	err = RegisterSingleton[Animal, lab](container, "buddy")
	assert.Nil(t, err, "error registering buddy depenedency")

	err = RegisterTransient[house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, _, err = GetContext[house](ctx)
	assert.NotNil(t, err, "no error returned for captive dependency")
	assert.Equal(t, "captive dependency error: buddy is a singleton but has a transient dependency john", err.Error())
}

func TestEnableCaptiveDependencies(t *testing.T) {
	type house struct {
		Owner Person `inject:""`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test enable captive dependencies",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterTransient[Person, Human](container)
	assert.Nil(t, err, "error registering owner depenedency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering house depenedency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseVal, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting house instance")

	assert.NotNil(t, houseVal.Owner, "owner dependency was not set")
	assert.Equal(t, 1, houseVal.Owner.Count())

	// get dep again
	_, houseVal, err = GetContext[house](ctx)
	assert.Nil(t, err, "error getting house instance")

	assert.NotNil(t, houseVal.Owner, "owner dependency was not set")
	assert.Equal(t, 2, houseVal.Owner.Count())
}

func TestEnableUnsafeDependencies(t *testing.T) {
	type house struct {
		dad Person `inject:""`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test enable unsafe dependencies",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  true,
	}

	// register deps
	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Person, Human](container)
	assert.Nil(t, err, "error registering dependency dad dependency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering dependency house dependency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseVal, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting test struct")

	assert.NotNil(t, houseVal.dad, "Dependency was not set")
}

func TestDisableUnsafeDependencies(t *testing.T) {
	type house struct {
		dad Person `inject:""`
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "test disable unsafe dependencies",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}

	// register deps
	container, err := NewDIContainer(config)
	assert.Nil(t, err, "error creating container")

	err = RegisterSingleton[Person, Human](container)
	assert.Nil(t, err, "error registering dependency dad dependency")

	err = RegisterSingleton[house, house](container)
	assert.Nil(t, err, "error registering dependency house dependency")

	ctx := context.Background()
	ctx, err = SetActiveContainer(ctx, config.ID)
	assert.Nil(t, err, "error setting active container")

	_, houseVal, err := GetContext[house](ctx)
	assert.Nil(t, err, "error getting test struct")

	assert.Nil(t, houseVal.dad, "Dependency was not set")
}
