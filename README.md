# ectoinject

ectoinject is a library that makes dependency injection simple and user friendly.
It's designed to make it painless to manage your applications dependencies while
remaining flexible.

Features:

- Easy dependency registration
- Multiple dependency lifecycles
- Complex type handling

## Install

`go get -u github.com/Gobusters/ectoinject`

## Concepts

### Lifecycles

**Singleton**:
A singleton dependency is created once and saved for the lifetime on the application. Singletons are the most performant dependencies but may cause issues if your dependency is stateful.

**Scoped**: A scoped dependency is created once per context.Context. This is useful for situations where you need temporary statefulness between dependencies.

**Transient**: A transient dependency is created everytime the dependency is requested.

**Captive Dependencies**: A Captive dependency occurs when a dependencies parent has a longer lifecycle than the dependency. For example, if we have dependency `foo` that is a singleton and has dependency on transient dependency `bar`. When `foo` is created, an instance of `bar` will be created, but because `foo` is a singleton, `bar` will be captive till `foo` is deleted.

## Usage

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
instance as the first return value, you can optionally provide a error as the second return value.

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
