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
