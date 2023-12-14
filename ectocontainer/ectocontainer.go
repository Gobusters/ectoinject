package ectocontainer

import (
	"context"

	"github.com/Gobusters/ectoinject/dependency"
)

// DIContainer is the interface for the container
type DIContainer interface {
	Get(ctx context.Context, name string) (context.Context, any, error) // Gets a dependency from the container
	GetConstructorFuncName() string                                     // Gets the name of the constructor function
	AddDependency(dep dependency.Dependency)                            // Adds a dependency to the container
	GetContainerID() string                                             // Gets the id of the container
}

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
