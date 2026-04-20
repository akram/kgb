package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	"github.com/akram/kgb/pkg/audit"
	"github.com/akram/kgb/pkg/drift"
	"github.com/akram/kgb/pkg/intent"
	"github.com/akram/kgb/pkg/policyengine"
	"github.com/akram/kgb/pkg/replay"
	"github.com/akram/kgb/pkg/trustmode"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Server wires HTTP handlers for evaluation, approvals, optional reverse proxy, and metrics.
type Server struct {
	Client       client.Client
	Analyzer     intent.Analyzer
	Engine       *policyengine.Engine
	Audit        *audit.Logger
	Approvals    *MemoryApprovals
	DefaultNS    string
	UpstreamURL  *url.URL
	TrustDefault trustmode.Mode
}

// EvaluateRequest is the JSON body for POST /v1/evaluate.
type EvaluateRequest struct {
	Namespace      string         `json:"namespace"`
	Session        string         `json:"session"`
	Payload        json.RawMessage `json:"payload"`
	SessionLabels  map[string]string `json:"sessionLabels,omitempty"`
	PreviousIntent *intent.Intent `json:"previousIntent,omitempty"`
}

// EvaluateResponse documents the decision returned to agents.
type EvaluateResponse struct {
	Intent       intent.Intent `json:"intent"`
	Outcome      string        `json:"outcome"`
	Reason       string        `json:"reason"`
	Policy       string        `json:"policy,omitempty"`
	Drift        bool          `json:"drift"`
	DriftReason  string        `json:"driftReason,omitempty"`
	Replay       bool          `json:"replay"`
}

func (s *Server) namespaceFrom(r *http.Request, bodyNS string) string {
	if v := strings.TrimSpace(bodyNS); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-KGB-Namespace")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("KGB_DEFAULT_NAMESPACE")); v != "" {
		return v
	}
	if s.DefaultNS != "" {
		return s.DefaultNS
	}
	return "default"
}

func (s *Server) trustFrom(r *http.Request) trustmode.Mode {
	if v := r.Header.Get("X-KGB-Trust-Mode"); v != "" {
		return trustmode.Parse(v)
	}
	return s.TrustDefault
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/v1/evaluate", s.handleEvaluate)
	mux.HandleFunc("/approvals/", s.handleApprovals)
	if s.UpstreamURL != nil {
		proxy := httputil.NewSingleHostReverseProxy(s.UpstreamURL)
		mux.Handle("/proxy/", http.StripPrefix("/proxy", proxy))
	}
	return withLogging(mux)
}

func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	in := intent.Input{
		Namespace:   s.namespaceFrom(r, req.Namespace),
		SessionName: req.Session,
		RawBody:     req.Payload,
	}
	it, err := s.Analyzer.Analyze(ctx, in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	driftHit, driftWhy := drift.Detect(req.PreviousIntent, it)

	mode := s.trustFrom(r)
	replayMode := replay.Enabled()

	var pr policyengine.Result
	if s.Engine != nil && !replayMode {
		var err error
		pr, err = s.Engine.Evaluate(ctx, in.Namespace, it, mode, req.SessionLabels)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		pr = policyengine.Result{Outcome: policyengine.OutcomeSurveillance, Reason: "replay_mode_stub"}
	}

	resp := EvaluateResponse{
		Intent:      it,
		Outcome:     string(pr.Outcome),
		Reason:      pr.Reason,
		Drift:       driftHit,
		DriftReason: driftWhy,
		Replay:      replayMode,
	}
	if pr.Policy != nil {
		resp.Policy = pr.Policy.Name
	}
	if s.Audit != nil {
		s.Audit.Write(audit.Entry{
			Namespace: in.Namespace,
			Session:   req.Session,
			Intent:    it,
			Outcome:   resp.Outcome,
			Policy:    resp.Policy,
			Reason:    resp.Reason,
			Replay:    replayMode,
		})
	}
	evaluations.WithLabelValues(resp.Outcome).Inc()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleApprovals(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/approvals/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	verb := strings.ToLower(parts[1])
	if verb != "allow" && verb != "deny" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Approvals == nil {
		http.Error(w, "approvals store not configured", http.StatusServiceUnavailable)
		return
	}
	switch verb {
	case "allow":
		s.Approvals.Set(id, string(kgbv1alpha1.ApprovalAllowed))
	case "deny":
		s.Approvals.Set(id, string(kgbv1alpha1.ApprovalDenied))
	}
	w.WriteHeader(http.StatusNoContent)
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start).String())
	})
}
