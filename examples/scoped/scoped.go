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
