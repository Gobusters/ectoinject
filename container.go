package ectoinject

import (
	"github.com/Gobusters/ectoinject/internal/container"
	"github.com/Gobusters/ectoinject/internal/logging"
	"github.com/Gobusters/ectoinject/loglevel"

	containermodels "github.com/Gobusters/ectoinject/container"
)

var defaultContainerID = "default"

var defaulLoggerConfig = containermodels.DIContainerLoggerConfig{
	Prefix:      "ectoinject",
	LogLevel:    loglevel.INFO,
	EnableColor: true,
	Enabled:     true,
	LogFunc:     nil,
}

// NewDIDefaultContainer creates a new container with the default configuration
func NewDIDefaultContainer() (DIContainer, error) {
	logger, err := logging.NewLogger(defaulLoggerConfig.Prefix, defaulLoggerConfig.LogLevel, defaulLoggerConfig.EnableColor, defaulLoggerConfig.Enabled, defaulLoggerConfig.LogFunc)
	if err != nil {
		return nil, err
	}

	ectoContainer := container.NewEctoContainer(containermodels.DIContainerConfig{
		ID:                       defaultContainerID,
		AllowCaptiveDependencies: true,
		AllowMissingDependencies: true,
		RequireInjectTag:         false,
		AllowUnsafeDependencies:  false,
		ConstructorFuncName:      "Constructor",
		InjectTagName:            "inject",
	}, logger)

	container.AddContainer(ectoContainer)

	return ectoContainer, nil
}

func NewDIContainer(config containermodels.DIContainerConfig) (DIContainer, error) {
	if config.ID == "" {
		config.ID = defaultContainerID
	}

	if config.LoggerConfig == nil {
		config.LoggerConfig = &defaulLoggerConfig
	}

	loggerConfig := config.LoggerConfig
	logger, err := logging.NewLogger(loggerConfig.Prefix, loggerConfig.LogLevel, loggerConfig.EnableColor, loggerConfig.Enabled, loggerConfig.LogFunc)
	if err != nil {
		return nil, err
	}

	ectoContainer := container.NewEctoContainer(config, logger)

	container.AddContainer(ectoContainer)

	return ectoContainer, nil
}
