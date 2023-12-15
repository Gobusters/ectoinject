# ectoinject

Simple, easy, user friendly Dependency Injection for Go projects. `ectoinject` is designed to allow users to quickly implement dependency injection without requiring extensive boilerplate code or code generation while still remaining flexible. Simply create a container, register the dependencies, and start using the dependencies.

Features:

- Easy dependency registration
- Multiple dependency lifecycles
- Complex type handling
- Unit testable
- Clear error messages

## Table of Contents

- [Install](#install)
- [Requirements](#requirements)
- [Concepts](#concepts)
  - [Lifecycles](#lifecycles)
- [Usage](#usage)
  - [Examples](#examples)
  - [Basic](#basic)
  - [Named Dependencies](#named-dependencies)
  - [Scoped Dependencies](#scoped-dependencies)
  - [Constructors](#constructors)
  - [Constructors with DIContainer dependency](#constructors-with-dicontainer-dependency)
  - [Instance Dependencies](#instance-dependencies)
  - [Custom Instance Getters](#custom-instance-getters)
- [Configuration](#configuration)
  - [AllowCaptiveDependencies](#allowcaptivedependencies)
  - [AllowMissingDependencies](#allowmissingdependencies)
  - [RequireInjectTag](#requireinjecttag)
  - [AllowUnsafeDependencies](#allowunsafedependencies)
  - [RequireConstructor](#requireconstructor)
  - [ConstructorFuncName](#constructorfuncname)
  - [InjectTagName](#injecttagname)
- [Logging](#logging)
  - [Prefix](#prefix)
  - [LogLevel](#loglevel)
  - [EnableColors](#enablecolors)
  - [Enabled](#enabled)
  - [Custom Logging](#custom-logging)
- [Inject Tag](#inject-tag)
- [Multiple Containers](#multiple-containers)
- [Unit Testing](#unit-testing)
- [Tips and Tricks](#tips-and-tricks)

## Install

`go get -u github.com/Gobusters/ectoinject`

## Requirements

```
go >= 1.18
```

## Concepts

### Lifecycles

**Singleton**:
A singleton dependency is created once and saved for the lifetime of the application. Singletons are the most performant dependencies but may cause issues if your dependency is stateful.

**Scoped**: A scoped dependency is created once per context.Context. This is useful for situations where you need temporary statefulness between dependencies.

**Transient**: A transient dependency is created everytime the dependency is requested.

**Captive Dependencies**: A Captive dependency occurs when a dependency's parent has a longer lifecycle than the dependency itself. For example, if we have dependency `foo` that is a singleton and has dependency on transient dependency `bar`. When `foo` is created, an instance of `bar` will be created, but because `foo` is a singleton, `bar` will be captive till `foo` is deleted.

## Usage

### Examples

Check out the example projects in the [examples folder](https://github.com/Gobusters/ectoinject/tree/main/examples)

### Basic

Simple implementation to get you started

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
)

type Bar struct {
}

func (b *Bar) Hello() string {
	return "Hello"
}

type Foo struct {
	Bar Bar
}

func main() {
	// create default container
	container, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register Bar as a singleton
	err = ectoinject.RegisterSingleton[Bar, Bar](container)
	if err != nil {
		panic(err) // handle error
	}

	// register Foo as a singleton
	err = ectoinject.RegisterSingleton[Foo, Foo](container)
	if err != nil {
		panic(err) // handle error
	}

	// create a new context with the container
	ctx := context.Background()

	// get Foo from the container
	ctx, foo, err := ectoinject.GetContext[Foo](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(foo.Bar.Hello()) // prints "Hello"
}
```

### Named Dependencies

Often its necessary to implement multiple concrete structs for a single interface. To differentiate
between them, we can give them names

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
)

type Person interface {
	Speak() string
}

type Peter struct {
}

func (p *Peter) Speak() string {
	return "We came, we saw, we kicked its ass!"
}

type Ray struct {
}

func (r *Ray) Speak() string {
	return "I tried to think of the most harmless thing. Something I loved from my childhood. Something that could never ever possibly destroy us. Mr. Stay Puft!"
}

type GhostBusters struct {
	Peter Person `inject:"peter"`
	Ray   Person `inject:"ray"`
}

func main() {
	// create default container
	container, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register Peter as a singleton
	err = ectoinject.RegisterSingleton[Person, Peter](container, "peter") // name needs to match the inject tag
	if err != nil {
		panic(err) // handle error
	}

	// register Peter as a singleton
	err = ectoinject.RegisterSingleton[Person, Ray](container, "ray") // name needs to match the inject tag
	if err != nil {
		panic(err) // handle error
	}

	// register GhostBusters as a singleton
	err = ectoinject.RegisterSingleton[GhostBusters, GhostBusters](container)
	if err != nil {
		panic(err) // handle error
	}

	// create a new context with the container
	ctx := context.Background()

	// get GhostBusters from the container
	ctx, gb, err := ectoinject.GetContext[GhostBusters](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.Peter.Speak()) // prints "We came, we saw, we kicked its ass!"
	println(gb.Ray.Speak())   // prints "I tried to think of the most harmless thing. Something I loved from my childhood. Something that could never ever possibly destroy us. Mr. Stay Puft!"
}
```

### Scoped Dependencies

Below is an example showing how you can utilze scoped dependencies

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
)

type GhostBuster interface {
	CaptureGhost()
	HasGhost() bool
}

type Ray struct {
	hasGhost bool
}

func (r *Ray) CaptureGhost() {
	r.hasGhost = true
}

func (r *Ray) HasGhost() bool {
	return r.hasGhost
}

func main() {
	// create default container
	container, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register Ray as Scoped
	err = ectoinject.RegisterScoped[GhostBuster, Ray](container)
	if err != nil {
		panic(err) // handle error
	}

	// create a new context with the container
	ctx := context.Background()

	// get Ray from the container. The context thats returned should be used for all subsequent calls that need the scoped Ray
	ctx, gb, err := ectoinject.GetContext[GhostBuster](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.HasGhost()) // false
	gb.CaptureGhost()
	println(gb.HasGhost()) // true

	// get Ray from the container again
	ctx, gb, err = ectoinject.GetContext[GhostBuster](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.HasGhost()) // true

	// get Ray from the container again but with a new context
	ctx = context.Background() // The new context will not have the scoped Ray. It will be a new instance of Ray
	ctx, gb, err = ectoinject.GetContext[GhostBuster](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.HasGhost()) // false
}
```

### Constructors

An alternative way to building a dependency is the use of the Constructor. On the dependency Struct, you may
define a method with the name "Constructor" (this name can be changed with [Configuration](##Configuration)).
You simply define your dependencies as arguments in the constructor method. The method should return the dependnecy
instance as the first return value, you can optionally provide a error as the second return value. **Note:** currently `ectoinject` does not support named dependencies for the constructor method. A work around is [Constructors with DIContainer dependency](###Constructors-with-DIContainer-dependency).

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
)

type GhostTrap struct {
	hasGhost bool
	isSet    bool
}

func (gt *GhostTrap) Set() {
	gt.isSet = true
}

func (gt *GhostTrap) IsSet() bool {
	return gt.isSet
}

func (gt *GhostTrap) Trigger() {
	gt.hasGhost = true
}

func (gt *GhostTrap) HasGhost() bool {
	return gt.hasGhost
}

type GhostBuster interface {
	CaptureGhost()
	HasGhost() bool
}

type Ray struct {
	ghostTrap *GhostTrap
	isReady   bool
}

func (r *Ray) CaptureGhost() {
	if r.isReady && r.ghostTrap.IsSet() {
		r.ghostTrap.Trigger()
	}
}

func (r *Ray) HasGhost() bool {
	return r.ghostTrap.HasGhost()
}

// Ray requires a GhostTrap to be injected and returns an instance of the Ray struct
func (r *Ray) Constructor(gt *GhostTrap) GhostBuster {
	r.isReady = true

	r.ghostTrap = gt
	r.ghostTrap.Set()

	return r
}

type GhostBusters struct {
	Ray GhostBuster `inject:"ray"`
}

func main() {
	// create default container
	container, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register GhostTrap as Singleton
	err = ectoinject.RegisterTransient[GhostTrap, GhostTrap](container)
	if err != nil {
		panic(err) // handle error
	}

	// register Ray as Singleton
	err = ectoinject.RegisterSingleton[GhostBuster, Ray](container, "ray")
	if err != nil {
		panic(err) // handle error
	}

	// register GhostBusters as Singleton
	err = ectoinject.RegisterSingleton[GhostBusters, GhostBusters](container)
	if err != nil {
		panic(err) // handle error
	}

	// create a new context with the container
	ctx := context.Background()

	// get GhostBusters from the container
	ctx, gb, err := ectoinject.GetContext[GhostBusters](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.Ray.HasGhost()) // false

	gb.Ray.CaptureGhost()

	println(gb.Ray.HasGhost()) // true
}
```

### Constructors with DIContainer dependency

You can also inject the DIContainer and context.Context into your constructor if you need

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
)

type GhostTrap struct {
	hasGhost bool
	isSet    bool
}

func (gt *GhostTrap) Set() {
	gt.isSet = true
}

func (gt *GhostTrap) IsSet() bool {
	return gt.isSet
}

func (gt *GhostTrap) Trigger() {
	gt.hasGhost = true
}

func (gt *GhostTrap) HasGhost() bool {
	return gt.hasGhost
}

type GhostBuster interface {
	CaptureGhost()
	HasGhost() bool
}

type Ray struct {
	ghostTrap *GhostTrap
	isReady   bool
}

func (r *Ray) CaptureGhost() {
	if r.isReady && r.ghostTrap.IsSet() {
		r.ghostTrap.Trigger()
	}
}

func (r *Ray) HasGhost() bool {
	return r.ghostTrap.HasGhost()
}

// Ray requires a GhostTrap to be injected and returns an instance of the Ray struct
func (r *Ray) Constructor(ctx context.Context, di ectoinject.DIContainer) GhostBuster {
	ctx, gt, err := ectoinject.GetContext[GhostTrap](ctx)
	if err != nil {
		panic(err) // handle error
	}
	r.isReady = true

	r.ghostTrap = &gt
	r.ghostTrap.Set()

	return r
}

type GhostBusters struct {
	Ray GhostBuster `inject:"ray"`
}

func main() {
	// create default container
	container, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register GhostTrap as Singleton
	err = ectoinject.RegisterTransient[GhostTrap, GhostTrap](container)
	if err != nil {
		panic(err) // handle error
	}

	// register Ray as Singleton
	err = ectoinject.RegisterSingleton[GhostBuster, Ray](container, "ray")
	if err != nil {
		panic(err) // handle error
	}

	// register GhostBusters as Singleton
	err = ectoinject.RegisterSingleton[GhostBusters, GhostBusters](container)
	if err != nil {
		panic(err) // handle error
	}

	// create a new context with the container
	ctx := context.Background()

	// get GhostBusters from the container
	ctx, gb, err := ectoinject.GetContext[GhostBusters](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.Ray.HasGhost()) // false

	gb.Ray.CaptureGhost()

	println(gb.Ray.HasGhost()) // true
}
```

### Instance Dependencies

You may encounter usecases where you wish to create a dependency instance that you share
throughout your project. That can be done using the `RegisterInstance` func.

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
)

type GhostBuster interface {
	WhoYaGonnaCall() string
}

type Peter struct {
	question string
}

func (r *Peter) WhoYaGonnaCall() string {
	return r.question
}

type Ray struct {
	Response string `inject:"catchphrase"`
}

func (r *Ray) WhoYaGonnaCall() string {
	return r.Response
}

type GhostBusters struct {
	Ray   GhostBuster `inject:"ray"`
	Peter GhostBuster `inject:"peter"`
}

func main() {
	// create default container
	container, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register Ray as Singleton
	err = ectoinject.RegisterSingleton[GhostBuster, Ray](container, "ray")
	if err != nil {
		panic(err) // handle error
	}

	// register the catchphrase as an instance with name "catchphrase"
	err = ectoinject.RegisterInstance[string](container, "Ghostbusters!", "catchphrase")
	if err != nil {
		panic(err) // handle error
	}

	peterInstance := &Peter{question: "Who ya gonna call?"} // create an instance of Peter
	// register peter as an instance with name "peter"
	err = ectoinject.RegisterInstance[GhostBuster](container, peterInstance, "peter")
	if err != nil {
		panic(err) // handle error
	}

	// register GhostBusters as Singleton
	err = ectoinject.RegisterSingleton[GhostBusters, GhostBusters](container)
	if err != nil {
		panic(err) // handle error
	}

	// create a new context with the container
	ctx := context.Background()

	// get GhostBusters from the container
	ctx, gb, err := ectoinject.GetContext[GhostBusters](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.Peter.WhoYaGonnaCall()) // prints "Who ya gonna call?"
	println(gb.Ray.WhoYaGonnaCall())   // prints "Ghostbusters!"
}
```

### Custom Instance Getters

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
	"github.com/Gobusters/ectoinject/lifecycles"
)

type GhostBuster interface {
	WhoYaGonnaCall() string
}

type Peter struct {
	question string
}

func (r *Peter) WhoYaGonnaCall() string {
	return r.question
}

type Ray struct {
	Response string `inject:"catchphrase"`
}

func (r *Ray) WhoYaGonnaCall() string {
	return r.Response
}

type GhostBusters struct {
	Ray   GhostBuster `inject:"ray"`
	Peter GhostBuster `inject:"peter"`
}

func main() {
	// create default container
	container, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register Ray as Singleton
	err = ectoinject.RegisterSingleton[GhostBuster, Ray](container, "ray")
	if err != nil {
		panic(err) // handle error
	}

	// register the catchphrase as an instance with name "catchphrase"
	err = ectoinject.RegisterInstance[string](container, "Ghostbusters!", "catchphrase")
	if err != nil {
		panic(err) // handle error
	}

	getPeter := func(ctx context.Context) (any, error) {
		return &Peter{question: "Who ya gonna call?"}, nil
	}

	// register peter as an instance with name "peter"
	err = ectoinject.RegisterInstanceFunc[GhostBuster](container, lifecycles.Singleton, getPeter, "peter")
	if err != nil {
		panic(err) // handle error
	}

	// register GhostBusters as Singleton
	err = ectoinject.RegisterSingleton[GhostBusters, GhostBusters](container)
	if err != nil {
		panic(err) // handle error
	}

	// create a new context with the container
	ctx := context.Background()

	// get GhostBusters from the container
	ctx, gb, err := ectoinject.GetContext[GhostBusters](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(gb.Peter.WhoYaGonnaCall()) // prints "Who ya gonna call?"
	println(gb.Ray.WhoYaGonnaCall())   // prints "Ghostbusters!"
}
```

## Configuration

`ectoinject` is intended to be flexible enough to handle your unique needs and requirements. When creating your container, you can provide a `ectocontainer.DIContainerConfig`. This allows you to change the behavior of the container as it gets dependencies.

```go
	config := ectocontainer.DIContainerConfig{
		ID:                       "my-custom-container",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
		RequireConstructor:       false,
		ConstructorFuncName:      "MyConstructorFunc",
		InjectTagName:            "MyInjectTag",
		LoggerConfig: &ectocontainer.DIContainerLoggerConfig{
			Prefix:      "ectoinject",
			LogLevel:    loglevel.INFO,
			EnableColor: true,
			Enabled:     true,
			LogFunc: func(ctx context.Context, level, msg string) {
				fmt.Printf("%s: %s\n", level, msg)
			},
		},
	}

	container, err := NewDIContainer(config)
```

### AllowCaptiveDependencies

If enabled, `AllowCaptiveDependencies` will only log a message if a captive dependency is detected. If false, an error will be returned.

### AllowMissingDependencies

If enabled, `AllowMissingDependencies` will ignore dependencies that have not been registered. If false, an error will be returned if a dependency is not found.

### RequireInjectTag

If enabled, `RequireInjectTag` struct fields without the inject tag will be ignored. If false, the container will attempt to inject a dependency for all struct fields.

### AllowUnsafeDependencies

If enabled, the container will attempt to inject a dependency for non-exported struct fields. If false, the container will ignore non-exported struct fields. Below is an example where `Foo` has a dependency on an unexported dependency `bar`. If `AllowUnsafeDependencies` is enabled, the container will attempt to inject `bar` despite it not being exported

```go
type Foo struct {
	bar Bar // unexported field
}
```

### RequireConstructor

If enabled, the container will only inject dependencies via the constructor function.

### ConstructorFuncName

Defines the name of the constructor function the container will look to use. Defaults to "Constructor"

### InjectTagName

Defines the name of the inject tag on the struct. Defaults to "inject"

## Inject Tag

The inject tag allows the you specify a named dependency to be injected into your struct. The tag name used can be changed using the container configuration [InjectTagName](##InjectTagName). You can tell the container to ignore the field by giving it a name of "-".

```go
type Foo struct {
	Bar       Bar `inject:"bar"` // the container will look for dependency with name "bar"
	MyPrivate Dep `inject:"-"`   // the container will ignore this depenency
}
```

## Logging

`ectoinject` does log messages to stdout in some instances. These logs are intended to help you identify potential issues in the dependency tree. You can change the behavior of these logs using the `LoggerConfig` field on the [container configuration](##Configuration)

### Prefix

This is a string added to the front of each log message. It defaults to "ectoinject" and is intended to help you identify where the log message originated.

### LogLevel

This affects the verbosity of the logs. It defaults to `loglevel.INFO` but can be changed to `loglevel.WARN`.

### EnableColors

If enabled, the logs will be colored to indicate their log level. This is intended to help draw attention to specific messages.

### Enabled

If false, no logs will be output

### Custom Logging

Using the `LogFunc` you can override the logger behavior. This allows you set a custom logging function. It can be useful for situations where
you have a standardized logger or log format. The example below demonstrates implementing a [logrus](https://github.com/sirupsen/logrus) logger

```go
	myLogrusLogger := func(ctx context.Context, level, msg string) {
		if level == loglevel.INFO {
			logrus.WithContext(ctx).Info(msg)
		}
		if level == loglevel.WARN {
			logrus.WithContext(ctx).Warn(msg)
		}
	}
	config := ectoinject.DefaultContainerConfig
	config.LoggerConfig.LogFunc = myLogrusLogger
	// create default container
	container, err := ectoinject.NewDIContainer(config)
```

## Multiple Containers

Most usecases will only require a single default container, but `ectoinject` does support multi container usecases. You can toggle between containers by using `ctx, err = ectoinject.SetActiveContainer(ctx, "my container id")`. If you don't set an active container, the default is used. You can change the default container by using `err = ectoinject.SetDefaultContainer("my container id")`

```go
package main

import (
	"context"

	"github.com/Gobusters/ectoinject"
	"github.com/Gobusters/ectoinject/ectocontainer"
)

type MyInterface interface {
	Output() string
}

type Bar struct {
}

func (b *Bar) Output() string {
	return "Hello"
}

type Foo struct {
}

func (b *Foo) Output() string {
	return "World"
}

func main() {
	// create default container
	defaultContainer, err := ectoinject.NewDIDefaultContainer()
	if err != nil {
		panic(err) // handle error
	}

	// register Bar as a singleton in default container
	err = ectoinject.RegisterSingleton[MyInterface, Bar](defaultContainer)
	if err != nil {
		panic(err) // handle error
	}

	config := ectocontainer.DIContainerConfig{
		ID:                       "my custom container",
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
	}
	// create a new custom container with the config
	customContainer, err := ectoinject.NewDIContainer(config)

	// register Foo as a singleton in the custom container
	err = ectoinject.RegisterSingleton[MyInterface, Foo](customContainer)
	if err != nil {
		panic(err) // handle error
	}

	// create ctx. If no container is passed, default container is used
	ctx := context.Background()

	// get MyInterface from the container
	ctx, bar, err := ectoinject.GetContext[MyInterface](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(bar.Output()) // prints "Hello"

	// set the custom container in the context
	ctx, err = ectoinject.SetActiveContainer(ctx, config.ID)
	if err != nil {
		panic(err) // handle error
	}

	// get MyInterface from the container
	ctx, foo, err := ectoinject.GetContext[MyInterface](ctx)
	if err != nil {
		panic(err) // handle error
	}

	println(foo.Output()) // prints "World"
}
```

## Unit Testing

While its not recommended to directly access the dependency container within code you intend to unit, it is possible to mock the container for unit testing.

Code under test:

```go
package unittesting

import (
	"context"

	"github.com/Gobusters/ectoinject"
)

type Foo interface {
	Bar() string
}

func MyFunc(ctx context.Context) string {
	// code directly access the container
	_, foo, err := ectoinject.GetContext[Foo](ctx)
	if err != nil {
		panic(err) // handle error
	}

	return foo.Bar()
}
```

unit test with container mock

```go
package unittesting

import (
	"context"
	"testing"

	"github.com/Gobusters/ectoinject"
	"github.com/Gobusters/ectoinject/dependency"
)

type ContainerMock struct {
	FooMock Foo
	ID      string
}

func (m *ContainerMock) Get(ctx context.Context, name string) (context.Context, any, error) {
	return ctx, m.FooMock, nil
}

func (m *ContainerMock) GetConstructorFuncName() string {
	return ""
}

func (m *ContainerMock) AddDependency(dep dependency.Dependency) {

}

func (m *ContainerMock) GetContainerID() string {
	return m.ID
}

type FooMock struct {
}

func (m *FooMock) Bar() string {
	return "Hello World"
}

func TestWithMockContainer(t *testing.T) {
	containerID := "my mock container"
	fooMock := &FooMock{}
	containerMock := &ContainerMock{
		FooMock: fooMock,
		ID:      containerID,
	}

	err := ectoinject.RegisterContainer(containerMock)
	if err != nil {
		t.Fatalf("error registering container: %s", err)
	}

	ctx := context.Background()
	ctx, err = ectoinject.SetActiveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("error setting active container: %s", err)
	}

	result := MyFunc(ctx)
	if result != "Hello World" {
		t.Fatalf("expected result to be 'Hello World', got '%s'", result)
	}
}
```

## Tips and Tricks

### Stick to singletons

Singletons will provide you the best overall performance. They are created once per container and saved for the
lifetime of the application.

### Avoid Captive Dependencies

While captive dependencies do not break `ectoinject` they can make your code behave in ways you might not expect.

### Use interfaces

`ectoinject` fully supports interfaces and interfaces will help make your code more flexible and testable.
