package policyengine

import (
	"regexp"
	"strings"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	"github.com/akram/kgb/pkg/intent"
)

// MatchesIntent returns true when policy intentMatch + matchLabels fit.
func MatchesIntent(p kgbv1alpha1.Policy, it intent.Intent, sessionLabels map[string]string) bool {
	if p.Spec.IntentMatch == nil {
		return false
	}
	m := p.Spec.IntentMatch
	if m.Action != "" && m.Action != it.Action {
		return false
	}
	if m.TargetExact != "" && m.TargetExact != it.Target {
		return false
	}
	if m.TargetRegex != "" {
		re, err := regexp.Compile(m.TargetRegex)
		if err != nil || !re.MatchString(it.Target) {
			return false
		}
	}
	if m.TargetDomainSuffix != "" {
		d := domainOf(it.Target)
		if d == "" || !strings.HasSuffix(strings.ToLower(d), strings.ToLower(m.TargetDomainSuffix)) {
			return false
		}
	}
	if len(m.RiskLevels) > 0 {
		ok := false
		for _, rl := range m.RiskLevels {
			if strings.EqualFold(string(rl), it.Risk) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	for k, want := range p.Spec.MatchLabels {
		got, exists := sessionLabels[k]
		if !exists || got != want {
			return false
		}
	}
	return true
}

func domainOf(addr string) string {
	addr = strings.TrimSpace(strings.ToLower(addr))
	if i := strings.LastIndex(addr, "@"); i >= 0 && i+1 < len(addr) {
		return addr[i+1:]
	}
	return ""
}
