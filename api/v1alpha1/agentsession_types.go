package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TrustMode configures baseline strictness for unmatched intents.
// +kubebuilder:validation:Enum=paranoid;balanced;permissive
type TrustMode string

const (
	TrustParanoid    TrustMode = "paranoid"
	TrustBalanced    TrustMode = "balanced"
	TrustPermissive  TrustMode = "permissive"
)

// AgentSessionSpec identifies an agent and tenant context.
type AgentSessionSpec struct {
	AgentID   string    `json:"agentId"`
	TrustMode TrustMode `json:"trustMode,omitempty"`
	TenantID  string    `json:"tenantId,omitempty"`
}

// AgentSessionStatus captures last evaluated intent for drift checks.
type AgentSessionStatus struct {
	LastEvaluatedIntent *IntentSpec `json:"lastEvaluatedIntent,omitempty"`
	DriftDetected       bool        `json:"driftDetected,omitempty"`
	DriftReason         string      `json:"driftReason,omitempty"`
	LastTransitionTime  *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kgbas

// AgentSession models a running agent session for drift detection and policy scoping.
type AgentSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSessionSpec   `json:"spec,omitempty"`
	Status AgentSessionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentSessionList contains a list of AgentSession.
type AgentSessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentSession `json:"items"`
}
