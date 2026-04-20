package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntentSpec is structured intent attached to an approval or session.
type IntentSpec struct {
	Action      string    `json:"action"`
	Target      string    `json:"target,omitempty"`
	Risk        RiskLevel `json:"risk,omitempty"`
	Confidence  float64   `json:"confidence,omitempty"`
	RawPayload  string    `json:"rawPayload,omitempty"`
}

// ApprovalPhase tracks human or automated resolution.
// +kubebuilder:validation:Enum=Pending;Allowed;Denied;Expired
type ApprovalPhase string

const (
	ApprovalPending ApprovalPhase = "Pending"
	ApprovalAllowed ApprovalPhase = "Allowed"
	ApprovalDenied  ApprovalPhase = "Denied"
	ApprovalExpired ApprovalPhase = "Expired"
)

// ApprovalRequestSpec binds an intent snapshot to a session for audit.
type ApprovalRequestSpec struct {
	SessionRef *corev1.LocalObjectReference `json:"sessionRef,omitempty"`
	Intent     IntentSpec              `json:"intent"`
	Reason     string                  `json:"reason,omitempty"`
}

// ApprovalRequestStatus records workflow outcome.
type ApprovalRequestStatus struct {
	Phase       ApprovalPhase `json:"phase,omitempty"`
	Message     string        `json:"message,omitempty"`
	DecidedAt   *metav1.Time  `json:"decidedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kgbapr

// ApprovalRequest requests human or automated approval for a single intent.
type ApprovalRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApprovalRequestSpec   `json:"spec,omitempty"`
	Status ApprovalRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApprovalRequestList contains a list of ApprovalRequest.
type ApprovalRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApprovalRequest `json:"items"`
}
