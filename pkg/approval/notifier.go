package approval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
)

// Notifier surfaces async human notification channels.
type Notifier interface {
	NotifyPending(ctx context.Context, ar *kgbv1alpha1.ApprovalRequest) error
}

// MultiNotifier fans out to all configured backends.
type MultiNotifier []Notifier

// NotifyPending implements Notifier.
func (m MultiNotifier) NotifyPending(ctx context.Context, ar *kgbv1alpha1.ApprovalRequest) error {
	for _, n := range m {
		if n == nil {
			continue
		}
		if err := n.NotifyPending(ctx, ar); err != nil {
			return err
		}
	}
	return nil
}

// SlackWebhook posts a compact JSON payload when SLACK_WEBHOOK_URL is set.
type SlackWebhook struct {
	URL string
	RT  *http.Client
}

// NewSlackFromEnv returns nil if SLACK_WEBHOOK_URL is unset.
func NewSlackFromEnv() *SlackWebhook {
	u := strings.TrimSpace(os.Getenv("SLACK_WEBHOOK_URL"))
	if u == "" {
		return nil
	}
	return &SlackWebhook{URL: u, RT: http.DefaultClient}
}

// NotifyPending implements Notifier.
func (s *SlackWebhook) NotifyPending(ctx context.Context, ar *kgbv1alpha1.ApprovalRequest) error {
	if s == nil || s.URL == "" {
		return nil
	}
	payload := map[string]any{
		"text": fmt.Sprintf("KGB approval required: %s/%s intent=%s target=%s",
			ar.Namespace, ar.Name, ar.Spec.Intent.Action, ar.Spec.Intent.Target),
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.RT.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook status %s", resp.Status)
	}
	return nil
}

// RESTCallback posts the ApprovalRequest JSON to APPROVAL_CALLBACK_URL.
type RESTCallback struct {
	URL string
	RT  *http.Client
}

// NewRESTCallbackFromEnv may return nil.
func NewRESTCallbackFromEnv() *RESTCallback {
	u := strings.TrimSpace(os.Getenv("APPROVAL_CALLBACK_URL"))
	if u == "" {
		return nil
	}
	return &RESTCallback{URL: u, RT: &http.Client{Timeout: 10 * time.Second}}
}

// NotifyPending implements Notifier.
func (r *RESTCallback) NotifyPending(ctx context.Context, ar *kgbv1alpha1.ApprovalRequest) error {
	if r == nil || r.URL == "" {
		return nil
	}
	body, err := json.Marshal(ar)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.RT.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("approval callback status %s", resp.Status)
	}
	return nil
}
