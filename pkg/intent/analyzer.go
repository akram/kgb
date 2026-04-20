package intent

import "context"

// Analyzer turns raw request context into structured Intent.
type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, in Input) (Intent, error)
}

// Input carries everything classifiers may need (HTTP body, headers, session hints).
type Input struct {
	Namespace   string            `json:"namespace"`
	SessionName string            `json:"sessionName,omitempty"`
	RawBody     []byte            `json:"-"`
	Headers     map[string]string `json:"headers,omitempty"`
	Method      string            `json:"method,omitempty"`
	Path        string            `json:"path,omitempty"`
}
