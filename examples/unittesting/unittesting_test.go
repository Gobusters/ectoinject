package unittesting

import (
	"context"
	"testing"

	"github.com/Gobusters/ectoinject"
	"github.com/Gobusters/ectoinject/dependency"
)

type ContainerMock struct {
	FooMock Foo
	ID      string
}

func (m *ContainerMock) Get(ctx context.Context, name string) (context.Context, any, error) {
	return ctx, m.FooMock, nil
}

func (m *ContainerMock) GetConstructorFuncName() string {
	return ""
}

func (m *ContainerMock) AddDependency(dep dependency.Dependency) {

}

func (m *ContainerMock) GetContainerID() string {
	return m.ID
}

type FooMock struct {
}

func (m *FooMock) Bar() string {
	return "Hello World"
}

func TestWithMockContainer(t *testing.T) {
	containerID := "my mock container"
	fooMock := &FooMock{}
	containerMock := &ContainerMock{
		FooMock: fooMock,
		ID:      containerID,
	}

	err := ectoinject.RegisterContainer(containerMock)
	if err != nil {
		t.Fatalf("error registering container: %s", err)
	}

	ctx := context.Background()
	ctx, err = ectoinject.SetActiveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("error setting active container: %s", err)
	}

	result := MyFunc(ctx)
	if result != "Hello World" {
		t.Fatalf("expected result to be 'Hello World', got '%s'", result)
	}
}
