package trustmode

import (
	"os"
	"strings"
)

// Mode influences defaults when no Policy matches.
type Mode string

const (
	Paranoid   Mode = "paranoid"
	Balanced   Mode = "balanced"
	Permissive Mode = "permissive"
)

// Parse normalizes user-facing strings.
func Parse(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(Permissive):
		return Permissive
	case string(Paranoid):
		return Paranoid
	default:
		return Balanced
	}
}

// FromEnv reads KGB_TRUST_MODE.
func FromEnv() Mode {
	return Parse(os.Getenv("KGB_TRUST_MODE"))
}
