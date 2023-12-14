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
