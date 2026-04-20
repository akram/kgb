# Example: Agent Sends External Email

This walkthrough matches the user story: agent attempts `send_email` to an external address, KGB blocks, operator approves, a narrow policy is created.

## 1. Agent session (balanced trust mode)

```yaml
apiVersion: kgb.io/v1alpha1
kind: AgentSession
metadata:
  name: demo-session
  namespace: team-a
spec:
  agentId: support-bot-1
  trustMode: balanced
  tenantId: team-a
```

## 2. Baseline policy — internal only

```yaml
apiVersion: kgb.io/v1alpha1
kind: Policy
metadata:
  name: email-internal-only
  namespace: team-a
spec:
  priority: 100
  intentMatch:
    action: send_email
    targetDomainSuffix: company.com
  decision: allow
```

No policy matches `external_user@gmail.com` → **deny-by-default** or **require_approval** depending on trust mode. With `balanced`, KGB returns **require_approval** and creates an `ApprovalRequest`.

## 3. ApprovalRequest (created by gateway/controller)

```yaml
apiVersion: kgb.io/v1alpha1
kind: ApprovalRequest
metadata:
  name: ar-20260420-001
  namespace: team-a
spec:
  sessionRef:
    name: demo-session
  intent:
    action: send_email
    target: external_user@gmail.com
    risk: high
    confidence: 0.92
  reason: no matching allow policy for external recipient
```

A Kubernetes **Event** is emitted on the `ApprovalRequest`; Slack webhook fires if configured.

## 4. Human approves

```bash
curl -X POST http://kgb-gateway.team-a.svc/approvals/ar-20260420-001/allow
```

Controller marks `ApprovalRequest` as allowed and (optionally) creates:

```yaml
apiVersion: kgb.io/v1alpha1
kind: Policy
metadata:
  name: generated-allow-gmail-onetime
  namespace: team-a
spec:
  priority: 200
  intentMatch:
    action: send_email
    targetExact: external_user@gmail.com
  decision: time_bound_allow
  timeBoundAllow: 15m
```

## 5. Drift detection

If the next call changes intent to `send_email` → `ceo@company.com` with escalated risk, gateway compares to `AgentSession.status.lastEvaluatedIntent` and may open a **new** `ApprovalRequest`.
