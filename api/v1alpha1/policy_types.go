package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RiskLevel approximates OWASP-style coarse risk for intent.
// +kubebuilder:validation:Enum=low;medium;high;critical
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// PolicyDecision is the enforcement outcome when a policy matches.
// +kubebuilder:validation:Enum=deny;allow;require_approval;session_trust;time_bound_allow;surveillance
type PolicyDecision string

const (
	PolicyDeny             PolicyDecision = "deny"
	PolicyAllow            PolicyDecision = "allow"
	PolicyRequireApproval  PolicyDecision = "require_approval"
	PolicySessionTrust     PolicyDecision = "session_trust"
	PolicyTimeBoundAllow   PolicyDecision = "time_bound_allow"
	PolicySurveillance     PolicyDecision = "surveillance"
)

// TrustModeOverride narrows how aggressive default denials are for this policy row.
// +kubebuilder:validation:Enum=paranoid;balanced;permissive
type TrustModeOverride string

// IntentMatch selects intents that should trigger this policy.
type IntentMatch struct {
	Action               string      `json:"action,omitempty"`
	TargetExact          string      `json:"targetExact,omitempty"`
	TargetRegex          string      `json:"targetRegex,omitempty"`
	TargetDomainSuffix   string      `json:"targetDomainSuffix,omitempty"`
	RiskLevels           []RiskLevel `json:"riskLevels,omitempty"`
}

// PolicySpec defines matcher + decision.
type PolicySpec struct {
	Priority            int32               `json:"priority,omitempty"`
	IntentMatch         *IntentMatch        `json:"intentMatch,omitempty"`
	MatchLabels         map[string]string   `json:"matchLabels,omitempty"`
	Decision            PolicyDecision      `json:"decision"`
	TimeBoundAllow      string              `json:"timeBoundAllow,omitempty"`
	TrustModeOverride   TrustModeOverride   `json:"trustModeOverride,omitempty"`
}

// PolicyStatus is written by the controller after validation.
type PolicyStatus struct {
	ObservedGeneration int64  `json:"observedGeneration,omitempty"`
	Message            string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kgbpol

// Policy is a namespaced allow/deny/surveillance rule evaluated against intent.
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy.
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
