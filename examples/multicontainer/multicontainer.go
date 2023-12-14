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
