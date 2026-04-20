# KGB — Kubernetes Guard for Bots

KGB is a **Kubernetes-native** surveillance and policy gateway for AI agents. It sits between agents and the outside world, classifying **intent**, enforcing **policies** (CRDs), and driving **approval workflows** when needed.

## Documentation

- [Architecture](docs/architecture.md) — system diagram, data flow, security model
- [Example: email block + approval](docs/example-email-flow.md)

## Quick Start (Development)

```bash
make generate   # deepcopy / CRD generation (requires controller-gen via go run)
make test
make build
```

Install CRDs into a cluster:

```bash
kubectl apply -f deploy/crds/
```

Run gateway (evaluate API only; configure `KUBECONFIG`):

```bash
./bin/kgb-gateway --bind=:8080
```

Run controller:

```bash
./bin/kgb-controller --metrics-bind-address=:8081
```

## Configuration (Gateway)

| Env | Meaning |
|-----|---------|
| `KGB_DEFAULT_NAMESPACE` | Namespace for policy CR lookups when no `X-KGB-Namespace` header |
| `KGB_TRUST_MODE` | Default trust mode: `paranoid`, `balanced`, `permissive` |
| `KGB_REPLAY_MODE` | `true` — use replay/simulation backend (no live side effects) |
| `KGB_UPSTREAM_URL` | Optional reverse proxy upstream base URL |

## API Sketch

- `POST /v1/evaluate` — classify intent + policy decision (see `pkg/gateway`)
- `POST /approvals/{id}/allow` / `deny` — approval callbacks (stub store; wire to `ApprovalRequest` in cluster)

## Repository Layout

```
api/v1alpha1/          # Policy, ApprovalRequest, AgentSession types
cmd/kgb-controller/    # Operator
cmd/kgb-gateway/      # HTTP gateway
deploy/crds/          # YAML CRDs
deploy/rbac/          # Sample RBAC
internal/controller/ # Reconcilers
pkg/                  # Intent, policy, approval, audit, trust modes
```

## License

Apache-2.0

## Clone

```bash
git clone git@github.com:akram/kgb.git
```
