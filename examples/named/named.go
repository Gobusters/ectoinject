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
