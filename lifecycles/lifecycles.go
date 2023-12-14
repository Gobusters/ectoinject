package lifecycles

const (
	Transient = "transient" // Transient dependencies are not cached. A new instance is created every time the dependency is requested
	Singleton = "singleton" // Singleton dependencies are cached for the lifetime of the application. One instance is created and shared.
	Scoped    = "scoped"    // Scoped dependencies are cached for the lifetime of the context. One instance is created and shared per context.
)

var Lifecycles = []string{Transient, Singleton, Scoped}

// IsValid checks if the lifecycle is valid
func IsValid(lifecycle string) bool {
	for _, l := range Lifecycles {
		if l == lifecycle {
			return true
		}
	}

	return false
}
