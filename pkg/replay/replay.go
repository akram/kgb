package replay

import (
	"os"
	"strings"
)

// Enabled returns true when KGB_REPLAY_MODE simulates decisions without mutating cluster state.
func Enabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("KGB_REPLAY_MODE")))
	return v == "1" || v == "true" || v == "yes"
}
