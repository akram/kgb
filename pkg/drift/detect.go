package drift

import (
	"github.com/akram/kgb/pkg/intent"
)

// Detect returns true if the new intent meaningfully diverges from the previous snapshot.
func Detect(prev *intent.Intent, next intent.Intent) (bool, string) {
	if prev == nil {
		return false, ""
	}
	if prev.Action != next.Action {
		return true, "action_changed"
	}
	if prev.Target != "" && next.Target != "" && prev.Target != next.Target {
		return true, "target_changed"
	}
	if prev.Risk != "" && next.Risk != "" {
		if riskRank(next.Risk) > riskRank(prev.Risk) {
			return true, "risk_escalation"
		}
	}
	return false, ""
}

func riskRank(r string) int {
	switch r {
	case "low":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	case "critical":
		return 4
	default:
		return 0
	}
}
