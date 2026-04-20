# KGB — Kubernetes Guard for Bots

KGB is a **cluster-level security gateway** and **policy control plane** for AI agents. It enforces **deny-by-default** decisions on tool and API invocations, with optional human-in-the-loop approval backed by Kubernetes APIs and events.

## Goals

- Intercept agent traffic (HTTP first; gRPC/MCP as extensions).
- Classify **intent** (action, target, risk) via pluggable analyzers.
- Evaluate **policies** stored as CRDs (`Policy`, `ApprovalRequest`, `AgentSession`).
- Emit **Kubernetes Events** and optional webhooks when approval is required.
- Support **multi-tenant** isolation via namespaces and label selectors.
- **Trust modes** (paranoid / balanced / permissive) tune default strictness.
- **Replay / simulation** mode for dry-running policies against recorded traffic.

## High-Level Architecture (ASCII)

```
                         ┌─────────────────────────────────────────────────────────┐
                         │                    Kubernetes Cluster                    │
  Agent / Runtime        │                                                         │
  (OpenAI-compat,        │   ┌──────────────┐         ┌─────────────────────────┐ │
  MCP client, etc.)      │   │ kgb-gateway  │         │   kgb-controller          │ │
        │                │   │  Deployment  │         │  (controller-runtime)    │ │
        │  HTTP/gRPC     │   │              │         │                         │ │
        └───────────────►│   │ ┌──────────┐ │         │ Reconcile:              │ │
                         │   │ │ Intent   │ │         │  - Policy               │ │
                         │   │ │ pipeline │ │         │  - ApprovalRequest      │ │
                         │   │ └────┬─────┘ │         │  - AgentSession         │ │
                         │   │      ▼       │         │                         │ │
                         │   │ ┌──────────┐ │         │ Writes status,          │ │
                         │   │ │ Policy   │ │◄───────►│ events, optional        │ │
                         │   │ │ engine   │ │  watch  │ approval notifications  │ │
                         │   │ └────┬─────┘ │         └───────────┬─────────────┘ │
                         │   │      ▼       │                     │               │
                         │   │ allow/deny │                     │               │
                         │   │ + audit    │                     ▼               │
                         │   └──────┬───────┘              ┌──────────────┐       │
                                │                    │ API Server   │       │
                                │ proxy/upstream     │ CRDs + Events│       │
                                ▼                    └──────────────┘       │
                         │   Upstream tools / APIs / model providers              │
                         └─────────────────────────────────────────────────────────┘

        Observability: Prometheus metrics (/metrics), structured audit log, K8s Events
```

## Data Flow (Evaluate)

1. Gateway receives a proxied or direct **evaluate** request (payload + session context).
2. **Intent** analyzers produce structured `Intent` (rule-based, LLM, or chain).
3. **Policy engine** loads `Policy` objects in the tenant namespace (deny if none match).
4. If decision is **require-approval**, gateway/controller creates `ApprovalRequest`, emits **Event**, notifies Slack/email/REST (configurable).
5. Human or automation calls **approval API**; controller records decision and may materialize a narrow **Policy** (session/time-bound).
6. **Drift**: on subsequent calls, gateway compares new intent to session baseline; controller may reopen approval.

## Components (This Repository)

| Path | Role |
|------|------|
| `cmd/kgb-controller` | Operator entrypoint |
| `cmd/kgb-gateway` | HTTP gateway + reverse proxy hooks |
| `api/v1alpha1` | CRD Go types |
| `deploy/crds` | Installable CRD manifests |
| `internal/controller` | Reconcilers |
| `pkg/intent` | Pluggable intent backends |
| `pkg/policyengine` | Deny-by-default matcher |
| `pkg/approval` | Notifications + approval store interface |
| `pkg/audit` | Decision audit trail |
| `pkg/trustmode` | Paranoid / balanced / permissive behavior |

## Security Posture

- **Deny-by-default** unless a `Policy` explicitly allows or surveillance is configured with permissive trust mode semantics.
- Secrets for webhooks live in Kubernetes Secrets, referenced by controller/gateway env or volume mounts (not in CR spec).
- Gateway authentication to API server uses **ServiceAccount** with least-privilege RBAC (see `deploy/rbac`).

## Future Work

- Native gRPC server + MCP framing parser.
- UI dashboard (Grafana dashboards + optional web console).
- OPA/Rego and WASM plugins for intent and policy.
