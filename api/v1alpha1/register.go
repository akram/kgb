package v1alpha1

func init() {
	SchemeBuilder.Register(
		&Policy{},
		&PolicyList{},
		&ApprovalRequest{},
		&ApprovalRequestList{},
		&AgentSession{},
		&AgentSessionList{},
	)
}
