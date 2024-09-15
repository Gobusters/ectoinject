package ectoinject

import (
	"context"
	"fmt"

	"github.com/Gobusters/ectoinject/ectocontainer"
	"github.com/Gobusters/ectoinject/internal/store"
)

type contextKey string

var contextContainerIDKey = contextKey("ectoinject-dependency-container-id")

// SetActiveContainer sets the active container in the context
// ctx: The context to set the active container in
// id: The id of the container to set as active
func SetActiveContainer(ctx context.Context, id string) (context.Context, error) {
	if c := store.GetContainer(id); c == nil {
		return ctx, fmt.Errorf("container with id '%s' does not exist", id)
	}

	ctx = context.WithValue(ctx, contextContainerIDKey, id)

	return ctx, nil
}

// GetActiveContainer gets the active container from the context
// ctx: The context to get the active container from
func GetActiveContainer(ctx context.Context) (ectocontainer.DIContainer, error) {
	id, _ := ctx.Value(contextContainerIDKey).(string)

	if id == "" {
		id = store.GetDefaultContainerID()
	}

	c := store.GetContainer(id)
	if c == nil {
		return nil, fmt.Errorf("container with id '%s' does not exist", id)
	}

	return c, nil
}
