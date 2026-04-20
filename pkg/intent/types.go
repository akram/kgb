package intent

// Intent is the structured output of intent detection.
type Intent struct {
	Action     string  `json:"action"`
	Target     string  `json:"target,omitempty"`
	Risk       string  `json:"risk,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}
