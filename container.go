package ectoinject

import (
	"fmt"

	"github.com/Gobusters/ectoinject/internal/logging"
	"github.com/Gobusters/ectoinject/loglevel"
)

var defaultContainerID = "default"

// DIContainerLoggerConfig is the configuration for the logger used by the container
type DIContainerLoggerConfig struct {
	Prefix      string                  // The prefix to use for the logger
	LogLevel    string                  // The log level to use for the logger
	EnableColor bool                    // Enables colors in the log messages
	Enabled     bool                    // Enables the logger
	LogFunc     func(level, msg string) // A custom log function to use
}

// DIContainerConfig is the configuration for the container
type DIContainerConfig struct {
	ID                       string                   // The id of the container
	AllowCaptiveDependencies bool                     // Allows dependencies with mismatched lifecycles. For example a Singleton that depends on a Transient will treat the transient as a Singleton
	AllowMissingDependencies bool                     // Allows dependencies to be missing
	RequireInjectTag         bool                     // Requires the inject tag to be present on dependencies
	AllowUnsafeDependencies  bool                     // Allows dependencies to be injected in an unsafe manner. This allows private fields to be injected
	LoggerConfig             *DIContainerLoggerConfig // The logger configuration to use
	ConstructorFuncName      string                   // The name of the constructor to use
	InjectTagName            string                   // The name of the inject tag to use
}

// Container for dependencies
type DIContainer struct {
	DIContainerConfig                       // The configuration for the container
	logger            *logging.Logger       // The logger to use
	container         map[string]Dependency // The container of dependencies
}

var defaulLoggerConfig = DIContainerLoggerConfig{
	Prefix:      "ectoinject",
	LogLevel:    loglevel.INFO,
	EnableColor: true,
	Enabled:     true,
	LogFunc:     nil,
}

// NewDIDefaultContainer creates a new container with the default configuration
func NewDIDefaultContainer() (*DIContainer, error) {
	logger, err := logging.NewLogger(defaulLoggerConfig.Prefix, defaulLoggerConfig.LogLevel, defaulLoggerConfig.EnableColor, defaulLoggerConfig.Enabled, defaulLoggerConfig.LogFunc)
	if err != nil {
		return nil, err
	}

	container := &DIContainer{
		DIContainerConfig: DIContainerConfig{
			ID:                       defaultContainerID,
			AllowCaptiveDependencies: true,
			AllowMissingDependencies: true,
			RequireInjectTag:         false,
			AllowUnsafeDependencies:  false,
			ConstructorFuncName:      "Constructor",
			InjectTagName:            "inject",
		},
		logger:    logger,
		container: make(map[string]Dependency),
	}

	addContainer(container)

	return container, nil
}

// NewDIContainer creates a new container
func NewDIContainer(config DIContainerConfig) (*DIContainer, error) {
	if config.ID == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	if container := getContainer(config.ID); container != nil {
		return nil, fmt.Errorf("container with id '%s' already exists", config.ID)
	}

	if config.LoggerConfig == nil {
		config.LoggerConfig = &defaulLoggerConfig
	}

	if config.ConstructorFuncName == "" {
		config.ConstructorFuncName = "Constructor"
	}

	if config.InjectTagName == "" {
		config.InjectTagName = "inject"
	}

	logger, err := logging.NewLogger(config.LoggerConfig.Prefix, config.LoggerConfig.LogLevel, config.LoggerConfig.EnableColor, config.LoggerConfig.Enabled, config.LoggerConfig.LogFunc)
	if err != nil {
		return nil, err
	}

	container := &DIContainer{
		DIContainerConfig: config,
		container:         make(map[string]Dependency),
		logger:            logger,
	}

	// if this is the first container, set it as the default
	if getContainerCount() == 0 {
		defaultContainerID = config.ID
	}

	// add the container to the store
	addContainer(container)

	return container, nil
}
