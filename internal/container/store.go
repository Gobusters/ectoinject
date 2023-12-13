package container

import "fmt"

var defaultContainerID = "default"

// singleton instance of the DIContainers
var containers = map[string]*EctoContainer{}

func AddContainer(container *EctoContainer) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	if _, ok := containers[container.ID]; ok {
		return fmt.Errorf("container with id '%s' already exists", container.ID)
	}

	if len(containers) == 0 {
		defaultContainerID = container.ID
	}

	containers[container.ID] = container

	return nil
}

func GetContainer(id string) *EctoContainer {
	if id == "" {
		return nil
	}

	container, ok := containers[id]
	if !ok {
		return nil
	}

	return container
}