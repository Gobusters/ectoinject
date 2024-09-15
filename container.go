package ectoinject

import (
	"github.com/Gobusters/ectoinject/ectocontainer"
	"github.com/Gobusters/ectoinject/internal/container"
	"github.com/Gobusters/ectoinject/internal/logging"
	"github.com/Gobusters/ectoinject/internal/store"
	"github.com/Gobusters/ectoinject/loglevel"
)

var DefaulLoggerConfig = ectocontainer.DIContainerLoggerConfig{
	Prefix:      "ectoinject",
	LogLevel:    loglevel.INFO,
	EnableColor: true,
	Enabled:     true,
	LogFunc:     nil,
}

var DefaultContainerConfig = ectocontainer.DIContainerConfig{
	ID:                       store.GetDefaultContainerID(),
	AllowCaptiveDependencies: true,
	AllowMissingDependencies: true,
	RequireInjectTag:         false,
	AllowUnsafeDependencies:  false,
	ConstructorFuncName:      "Constructor",
	InjectTagName:            "inject",
	LoggerConfig:             &DefaulLoggerConfig,
}

// NewDIDefaultContainer creates a new container with the default configuration. The default configuration is:
// ID: "default"
// AllowCaptiveDependencies: true
// AllowMissingDependencies: true
// RequireInjectTag: false
// AllowUnsafeDependencies: false
// ConstructorFuncName: "Constructor"
// InjectTagName: "inject"
func NewDIDefaultContainer() (ectocontainer.DIContainer, error) {
	return NewDIContainer(DefaultContainerConfig)
}

// NewDIContainer creates a new container with the provided configuration
// config: The configuration to use for the container
func NewDIContainer(config ectocontainer.DIContainerConfig) (ectocontainer.DIContainer, error) {
	if config.ID == "" {
		config.ID = store.GetDefaultContainerID()
	}

	if config.LoggerConfig == nil {
		config.LoggerConfig = &DefaulLoggerConfig
	}

	if config.InjectTagName == "" {
		config.InjectTagName = "inject"
	}

	if config.ConstructorFuncName == "" {
		config.ConstructorFuncName = "Constructor"
	}

	loggerConfig := config.LoggerConfig
	logger, err := logging.NewLogger(loggerConfig.Prefix, loggerConfig.LogLevel, loggerConfig.EnableColor, loggerConfig.Enabled, loggerConfig.LogFunc)
	if err != nil {
		return nil, err
	}

	ectoContainer := container.NewEctoContainer(config, logger)

	err = RegisterInstance[ectocontainer.DIContainer](ectoContainer, &ectoContainer)
	if err != nil {
		return nil, err
	}

	err = store.RegisterContainer(ectoContainer)
	if err != nil {
		return nil, err
	}

	return ectoContainer, nil
}

// RegisterContainer adds the container to the container lookup
func RegisterContainer(container ectocontainer.DIContainer) error {
	return store.RegisterContainer(container)
}

// SetDefaultContainer sets the default container to use
func SetDefaultContainer(containerID string) error {
	return store.SetDefaultContainer(containerID)
}

// GetDefaultContainer gets the default container
func GetDefaultContainer() ectocontainer.DIContainer {
	return store.GetDefaultContainer()
}

// GetContainer gets the container with the provided id
func GetContainer(id string) ectocontainer.DIContainer {
	return store.GetContainer(id)
}
