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
