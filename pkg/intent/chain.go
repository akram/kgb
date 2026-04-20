package intent

import (
	"context"
	"errors"
)

// Chain runs analyzers in order; the first non-empty high-confidence result wins,
// otherwise results are merged heuristically.
type Chain struct {
	Steps []Analyzer
}

// Name implements Analyzer.
func (c *Chain) Name() string {
	return "chain"
}

// Analyze implements Analyzer.
func (c *Chain) Analyze(ctx context.Context, in Input) (Intent, error) {
	if len(c.Steps) == 0 {
		return Intent{}, errors.New("intent.Chain: no analyzers configured")
	}
	var merged Intent
	for _, step := range c.Steps {
		out, err := step.Analyze(ctx, in)
		if err != nil {
			return Intent{}, err
		}
		if out.Action != "" && out.Confidence >= 0.85 {
			return out, nil
		}
		if merged.Action == "" && out.Action != "" {
			merged = out
		}
	}
	if merged.Action == "" {
		return c.Steps[len(c.Steps)-1].Analyze(ctx, in)
	}
	return merged, nil
}
