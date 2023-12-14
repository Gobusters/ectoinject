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
	Bar Bar `inject:"bar"`
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
