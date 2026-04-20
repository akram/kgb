package policyengine

import (
	"context"
	"fmt"
	"sort"
	"strings"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	"github.com/akram/kgb/pkg/intent"
	"github.com/akram/kgb/pkg/trustmode"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Outcome is the enforcement decision after policy + trust evaluation.
type Outcome string

const (
	OutcomeAllow             Outcome = "allow"
	OutcomeDeny              Outcome = "deny"
	OutcomeRequireApproval   Outcome = "require_approval"
	OutcomeSurveillance      Outcome = "surveillance"
)

// Result bundles evaluation output.
type Result struct {
	Outcome Outcome
	Policy  *kgbv1alpha1.Policy
	Reason  string
}

// Engine evaluates intents against Policy CRs in a namespace (multi-tenant boundary).
type Engine struct {
	Client client.Client
}

// Evaluate applies deny-by-default semantics with trust-mode fallbacks.
func (e *Engine) Evaluate(ctx context.Context, namespace string, it intent.Intent, mode trustmode.Mode, sessionLabels map[string]string) (Result, error) {
	if e.Client == nil {
		return Result{}, fmt.Errorf("policyengine: kubernetes client not configured")
	}
	var list kgbv1alpha1.PolicyList
	if err := e.Client.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		if apierrors.IsForbidden(err) {
			return Result{Outcome: OutcomeDeny, Reason: "policy_list_forbidden"}, nil
		}
		return Result{}, err
	}
	sort.SliceStable(list.Items, func(i, j int) bool {
		return list.Items[i].Spec.Priority > list.Items[j].Spec.Priority
	})
	for i := range list.Items {
		p := list.Items[i]
		if !MatchesIntent(p, it, sessionLabels) {
			continue
		}
		return mapDecision(&p), nil
	}
	return defaultForTrust(mode, it, "no_matching_policy"), nil
}

func mapDecision(p *kgbv1alpha1.Policy) Result {
	switch p.Spec.Decision {
	case kgbv1alpha1.PolicyDeny:
		return Result{Outcome: OutcomeDeny, Policy: p, Reason: "policy_deny"}
	case kgbv1alpha1.PolicyAllow, kgbv1alpha1.PolicySessionTrust, kgbv1alpha1.PolicyTimeBoundAllow:
		return Result{Outcome: OutcomeAllow, Policy: p, Reason: string(p.Spec.Decision)}
	case kgbv1alpha1.PolicyRequireApproval:
		return Result{Outcome: OutcomeRequireApproval, Policy: p, Reason: "policy_require_approval"}
	case kgbv1alpha1.PolicySurveillance:
		return Result{Outcome: OutcomeSurveillance, Policy: p, Reason: "policy_surveillance"}
	default:
		return Result{Outcome: OutcomeDeny, Policy: p, Reason: "unknown_policy_decision"}
	}
}

func defaultForTrust(mode trustmode.Mode, it intent.Intent, suffix string) Result {
	r := strings.ToLower(it.Risk)
	switch mode {
	case trustmode.Paranoid:
		return Result{Outcome: OutcomeDeny, Reason: "deny_by_default_" + suffix}
	case trustmode.Permissive:
		return Result{Outcome: OutcomeSurveillance, Reason: "permissive_default_" + suffix}
	default: // balanced
		if r == "high" || r == "critical" {
			return Result{Outcome: OutcomeRequireApproval, Reason: "balanced_high_risk_" + suffix}
		}
		return Result{Outcome: OutcomeDeny, Reason: "balanced_default_" + suffix}
	}
}
