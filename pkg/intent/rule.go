package intent

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
)

// RuleAnalyzer is a deterministic, fast path classifier for common agent tools.
type RuleAnalyzer struct{}

// Name implements Analyzer.
func (RuleAnalyzer) Name() string {
	return "rules"
}

// Analyze implements Analyzer.
func (RuleAnalyzer) Analyze(_ context.Context, in Input) (Intent, error) {
	if action, target := JSONToolCall(in.RawBody); action != "" {
		risk := "medium"
		if action == "send_email" && isExternalConsumerEmail(target) {
			risk = "high"
		}
		return Intent{
			Action:     action,
			Target:     target,
			Risk:       risk,
			Confidence: 0.88,
		}, nil
	}
	body := strings.ToLower(string(in.RawBody))
	if strings.Contains(body, "send_email") || strings.Contains(body, "\"tool\":\"email") || strings.Contains(body, "mailto:") {
		target := extractEmailTarget(body)
		risk := "medium"
		if isExternalConsumerEmail(target) {
			risk = "high"
		}
		return Intent{
			Action:     "send_email",
			Target:     target,
			Risk:       risk,
			Confidence: 0.9,
		}, nil
	}
	if strings.Contains(body, "http://") || strings.Contains(body, "https://") {
		return Intent{
			Action:     "http_call",
			Target:     extractFirstURL(body),
			Risk:       "medium",
			Confidence: 0.72,
		}, nil
	}
	return Intent{
		Action:     "unknown",
		Target:     in.Path,
		Risk:       "high",
		Confidence: 0.55,
	}, nil
}

func extractEmailTarget(s string) string {
	// naive parse for demo payloads
	if i := strings.Index(s, "@"); i > 0 {
		start := i - 1
		for start > 0 && (s[start] == '"' || s[start] == ' ' || s[start] == ':') {
			start--
		}
		start++
		end := i + 1
		for end < len(s) && (s[end] == '.' || s[end] >= 'a' && s[end] <= 'z' || s[end] >= 'A' && s[end] <= 'Z' || s[end] >= '0' && s[end] <= '9' || s[end] == '@' || s[end] == '-' || s[end] == '_') {
			end++
		}
		if end > i {
			return s[start:end]
		}
	}
	return ""
}

func isExternalConsumerEmail(addr string) bool {
	addr = strings.ToLower(addr)
	if !strings.Contains(addr, "@") {
		return false
	}
	domain := addr[strings.LastIndex(addr, "@")+1:]
	internal := []string{"company.com", "redhat.com", "corp.local"}
	for _, d := range internal {
		if strings.HasSuffix(domain, d) {
			return false
		}
	}
	return true
}

func extractFirstURL(s string) string {
	idx := strings.Index(s, "http")
	if idx < 0 {
		return ""
	}
	rest := s[idx:]
	end := strings.IndexAny(rest, " \"'\n\r\t,}")
	if end > 0 {
		return rest[:end]
	}
	return rest
}

// JSONToolCall tries to parse OpenAI-style tool_calls JSON fragments.
func JSONToolCall(payload []byte) (action string, target string) {
	var probe struct {
		Tool  string `json:"tool"`
		Name  string `json:"name"`
		Input any    `json:"input"`
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	if err := dec.Decode(&probe); err != nil {
		return "", ""
	}
	action = probe.Tool
	if action == "" {
		action = probe.Name
	}
	switch v := probe.Input.(type) {
	case map[string]any:
		if to, ok := v["to"].(string); ok {
			target = to
		}
	}
	return action, target
}
