package lifecycles

const (
	Transient = "transient"
	Singleton = "singleton"
	Scoped    = "scoped"
)

var Lifecycles = []string{Transient, Singleton, Scoped}

func IsValid(lifecycle string) bool {
	for _, l := range Lifecycles {
		if l == lifecycle {
			return true
		}
	}

	return false
}
