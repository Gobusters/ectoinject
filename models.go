package ectoinject

import (
	"context"

	"github.com/Gobusters/ectoinject/dependency"
)

type DIContainer interface {
	Get(ctx context.Context, name string) (context.Context, any, error)
	GetConstructorFuncName() string
	AddDependency(dep dependency.Dependency)
}
