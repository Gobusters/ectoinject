package store

import (
	"fmt"

	"github.com/Gobusters/ectoinject/ectocontainer"
)

var defaultContainerID = "default"
var defaultContainerSet = false

// singleton instance of the ectocontainer.DIContainers
var containers = map[string]ectocontainer.DIContainer{}

func RegisterContainer(container ectocontainer.DIContainer) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	if _, ok := containers[container.GetContainerID()]; ok {
		return fmt.Errorf("container with id '%s' already exists", container.GetContainerID())
	}

	if len(containers) == 0 && !defaultContainerSet {
		defaultContainerID = container.GetContainerID()
		defaultContainerSet = true
	}

	containers[container.GetContainerID()] = container

	return nil
}

func GetContainer(id string) ectocontainer.DIContainer {
	if id == "" {
		return nil
	}

	container, ok := containers[id]
	if !ok {
		return nil
	}

	return container
}

func GetDefaultContainer() ectocontainer.DIContainer {
	return GetContainer(defaultContainerID)
}

func SetDefaultContainer(containerID string) error {
	if containerID == "" {
		return fmt.Errorf("containerID cannot be empty")
	}

	if _, ok := containers[containerID]; !ok {
		return fmt.Errorf("container with id '%s' does not exist", containerID)
	}
	defaultContainerID = containerID
	defaultContainerSet = true

	return nil
}

func GetDefaultContainerID() string {
	return defaultContainerID
}
