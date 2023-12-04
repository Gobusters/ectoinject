package ectoinject

import (
	"fmt"
)

// singleton instance of the DIContainers
var containers = map[string]*DIContainer{}

func addContainer(container *DIContainer) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	if _, ok := containers[container.ID]; ok {
		return fmt.Errorf("container with id '%s' already exists", container.ID)
	}

	containers[container.ID] = container

	return nil
}

func getContainer(id string) *DIContainer {
	if id == "" {
		return nil
	}

	container, ok := containers[id]
	if !ok {
		return nil
	}

	return container
}

func getContainerCount() int {
	return len(containers)
}
