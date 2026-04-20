package policyengine

import (
	"testing"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	"github.com/akram/kgb/pkg/intent"
)

func TestMatchesIntentActionAndDomain(t *testing.T) {
	p := kgbv1alpha1.Policy{
		Spec: kgbv1alpha1.PolicySpec{
			IntentMatch: &kgbv1alpha1.IntentMatch{
				Action:             "send_email",
				TargetDomainSuffix: "company.com",
			},
		},
	}
	it := intent.Intent{Action: "send_email", Target: "ceo@company.com", Risk: "low"}
	if !MatchesIntent(p, it, nil) {
		t.Fatal("expected match")
	}
	it.Target = "ceo@gmail.com"
	if MatchesIntent(p, it, nil) {
		t.Fatal("expected no match for gmail")
	}
}

func TestMatchesIntentLabels(t *testing.T) {
	p := kgbv1alpha1.Policy{
		Spec: kgbv1alpha1.PolicySpec{
			IntentMatch: &kgbv1alpha1.IntentMatch{Action: "http_call"},
			MatchLabels: map[string]string{"team": "a"},
		},
	}
	it := intent.Intent{Action: "http_call"}
	if MatchesIntent(p, it, map[string]string{"team": "b"}) {
		t.Fatal("expected label mismatch")
	}
	if !MatchesIntent(p, it, map[string]string{"team": "a"}) {
		t.Fatal("expected label match")
	}
}
