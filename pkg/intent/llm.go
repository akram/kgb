package intent

import (
	"context"
	"errors"
)

// LLMAnalyzer is a stub for an LLM-backed classifier (OpenAI, vLLM, etc.).
// Wire HTTP client + prompt templates in a future change.
type LLMAnalyzer struct {
	Model string
}

// Name implements Analyzer.
func (l LLMAnalyzer) Name() string {
	return "llm"
}

// Analyze implements Analyzer.
func (l LLMAnalyzer) Analyze(_ context.Context, _ Input) (Intent, error) {
	if l.Model == "" {
		return Intent{}, errors.New("intent.LLMAnalyzer: model not configured")
	}
	// Placeholder: LLM path not enabled in skeleton builds.
	return Intent{}, errors.New("intent.LLMAnalyzer: not implemented")
}
