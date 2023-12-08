package ectoinject

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Gobusters/ectoinject/internal/cache"
)

var contextContainerIDKey = "ectoinject-dependency-container-id"
var contextScopedContainerIDKey = "ectoinject-dependency-Scoped-container-id"

// SetActiveContainer sets the active container in the context
// ctx: The context to set the active container in
// id: The id of the container to set as active
func SetActiveContainer(ctx context.Context, id string) (context.Context, error) {
	if container := getContainer(id); container == nil {
		return ctx, fmt.Errorf("container with id '%s' does not exist", id)
	}

	ctx = context.WithValue(ctx, contextContainerIDKey, id)

	return ctx, nil
}

// GetActiveContainer gets the active container from the context
// ctx: The context to get the active container from
func GetActiveContainer(ctx context.Context) (*EctoContainer, error) {
	id, _ := ctx.Value(contextContainerIDKey).(string)

	if id == "" {
		id = defaultContainerID
	}

	container := getContainer(id)
	if container == nil {
		return nil, fmt.Errorf("container with id '%s' does not exist", id)
	}

	return container, nil
}

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// scopeContext scopes the context to for use with scoped dependencies. To prevent memory leaks, the context must be canceled or UnscopeContext must be called when the scope is finished
// ctx: The context to scope
func scopeContext(ctx context.Context) context.Context {
	// create a new Scoped container id
	id := generateID()

	// set the Scoped container id in the context
	ctx = context.WithValue(ctx, contextScopedContainerIDKey, id)

	// listen for the context to be done
	go func() {
		<-ctx.Done()
		// remove the Scoped container from the cache
		cache.RemoveScopedCache(id)
	}()

	return ctx
}

// unscopeContext unscopes the context from a scoped dependency. This releases the cache for the scope and should be called when the scope is finished
// ctx: The context to unscope
func unscopeContext(ctx context.Context) context.Context {
	id := getScopedID(ctx)
	// remove the Scoped container id from the context
	ctx = context.WithValue(ctx, contextScopedContainerIDKey, "")

	// remove the Scoped container from the cache
	cache.RemoveScopedCache(id)

	return ctx
}

func getScopedID(ctx context.Context) string {
	id, _ := ctx.Value(contextScopedContainerIDKey).(string)
	return id
}
