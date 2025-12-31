# Automated FinOps Enforcer

**Active Cost Governance for Kubernetes**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.24+-blue.svg)](https://kubernetes.io)

## What This Is

A production-grade Kubernetes controller that automatically detects and pauses idle workloads in non-production environments, delivering measurable cost savings without human intervention.

**Core Principle:** Dashboards don't save money. Actions do.

## The Problem

Cloud cost overruns are rarely caused by malice. They're caused by:

- Dev and staging environments running 24×7
- Forgotten experimental deployments
- Accumulated non-prod workloads
- Zero accountability for idle infrastructure

By the time your Cost Explorer shows the damage, the money is already gone.

## The Solution

This controller:

1. **Detects** idle Kubernetes workloads with high confidence
2. **Acts** automatically but conservatively (scale-to-zero)
3. **Explains** every action it takes
4. **Reverses** easily via Slack interaction
5. **Measures** savings in real-time

## Architecture

```
┌──────────────┐
│ Kubernetes   │
│ Cluster      │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│ OpenCost         │
│ (Cost Metrics)   │
└──────┬───────────┘
       │
       ▼
┌────────────────────────┐
│ FinOps Enforcer        │
│ (Controller)           │
│                        │
│ - Fetch cost data      │
│ - Evaluate policies    │
│ - Safe enforcement     │
└──────┬─────────────────┘
       │
       ▼
┌──────────────────────┐
│ Actions              │
│ - Scale to zero      │
│ - Annotate resource  │
│ - Notify team        │
└──────────────────────┘
```

## Key Features

### Policy-Driven Governance

Define what "idle" means for your organization:

```yaml
policyName: non-prod-idle-gc
scope:
  namespaces:
    include:
      - dev-*
      - staging-*
    exclude:
      - prod

conditions:
  idleWindow: 48h
  minHourlyCost: 2.0
  trafficThreshold:
    requestsPerMinute: 0

actions:
  type: scaleToZero
  notify: slack
  reactivationAllowed: true
```

### Human-in-the-Loop

Every enforcement action triggers a Slack notification with one-click reactivation:

```
🚨 Idle Resource Paused

Namespace: dev-payments
Deployment: invoice-worker
Idle Duration: 72 hours
Estimated Monthly Savings: $180

⏯️ Reactivate | 📄 View Details
```

### Safety Guardrails

- **Namespace allowlisting** - Production is never touched by default
- **Cooldown windows** - Prevents flapping
- **Bounded actions** - Max resources per run
- **Dry-run mode** - Test policies safely
- **Audit trail** - Every action is logged

### Real-Time Metrics

- `finops_paused_resources_total` - Resources currently paused
- `finops_estimated_savings_usd` - Projected monthly savings
- `finops_policy_matches_total` - Policy evaluation results
- `finops_actions_taken_total` - Enforcement actions by type

## What This Is NOT

This project intentionally **does not**:

- ❌ Use ML for cost forecasting
- ❌ Replace your billing system
- ❌ Enforce globally across all namespaces
- ❌ Delete resources permanently
- ❌ Touch production by default

**Philosophy:** Conservative automation that preserves trust.

## Quick Start

### Prerequisites

- Kubernetes cluster (1.24+)
- OpenCost installed
- Slack webhook (optional)
- Prometheus (for metrics)

### Installation

```bash
# Install via Helm
helm repo add finops-enforcer https://charts.finops-enforcer.io
helm install finops-enforcer finops-enforcer/finops-enforcer \
  --namespace finops-system \
  --create-namespace \
  --set opencost.endpoint=http://opencost.opencost:9003 \
  --set slack.webhookURL=<your-webhook>

# Or via kubectl
kubectl apply -f https://raw.githubusercontent.com/yourusername/finops-enforcer/main/deploy/install.yaml
```

### Configuration

1. Create a policy file:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: dev-idle-gc
  namespace: finops-system
spec:
  scope:
    namespaces:
      include:
        - dev-*
        - staging-*
  conditions:
    idleWindow: 48h
    minHourlyCost: 2.0
  actions:
    type: scaleToZero
    notify: slack
EOF
```

2. Monitor enforcement:

```bash
kubectl logs -n finops-system deployment/finops-enforcer -f
```

## Use Cases

### Scenario 1: Forgotten Dev Environment

A developer spins up a feature branch environment for testing. After the feature merges, the environment is forgotten. After 48 hours of zero traffic, the enforcer:

1. Detects idle state
2. Scales deployments to zero
3. Notifies team in Slack
4. Saves ~$120/month

### Scenario 2: Weekend Non-Prod

Staging environments run 24×7 but only used Monday-Friday. The enforcer:

1. Detects weekend idle patterns
2. Auto-pauses Friday evening
3. Team reactivates Monday morning
4. Saves ~$400/month

### Scenario 3: Load Test Cleanup

After load testing, high-resource deployments are left running. The enforcer:

1. Detects abnormal cost + zero traffic
2. Flags for review
3. Auto-pauses after confirmation window
4. Saves ~$800/month

## Project Structure

```
.
├── cmd/
│   ├── controller/          # Main controller binary
│   └── cli/                 # finops-ctl CLI tool
├── pkg/
│   ├── controller/          # Reconciliation logic
│   ├── policy/              # Policy engine
│   ├── cost/                # OpenCost integration
│   ├── enforcement/         # Action execution
│   ├── metrics/             # Prometheus metrics
│   └── notifications/       # Slack integration
├── api/
│   └── v1alpha1/            # CRD definitions
├── config/
│   ├── crd/                 # Custom Resource Definitions
│   ├── rbac/                # RBAC manifests
│   ├── manager/             # Controller deployment
│   └── samples/             # Example policies
├── deploy/
│   ├── helm/                # Helm chart
│   └── manifests/           # Raw Kubernetes YAML
├── test/
│   ├── e2e/                 # End-to-end tests
│   └── integration/         # Integration tests
└── docs/
    ├── DESIGN.md            # Architecture deep-dive
    ├── POLICIES.md          # Policy reference
    └── RUNBOOK.md           # Operational guide
```

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Kind or Minikube (for local testing)

### Local Development

```bash
# Clone repository
git clone https://github.com/yourusername/finops-enforcer.git
cd finops-enforcer

# Install dependencies
go mod download

# Run tests
make test

# Build
make build

# Run locally (against current kubeconfig context)
make run

# Build Docker image
make docker-build IMG=finops-enforcer:dev
```

### Testing

```bash
# Unit tests
make test

# Integration tests (requires kind cluster)
make test-integration

# E2E tests
make test-e2e

# Coverage report
make coverage
```

## Configuration Reference

### Controller Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: finops-enforcer-config
  namespace: finops-system
data:
  config.yaml: |
    opencost:
      endpoint: http://opencost.opencost:9003
      timeout: 30s
    
    enforcement:
      dryRun: false
      maxActionsPerRun: 10
      cooldownWindow: 1h
    
    notifications:
      slack:
        enabled: true
        webhookURL: ${SLACK_WEBHOOK_URL}
        channel: "#finops-alerts"
    
    metrics:
      enabled: true
      port: 8080
      path: /metrics
```

### Policy Specification

See [POLICIES.md](docs/POLICIES.md) for complete reference.

## Metrics & Monitoring

### Key Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `finops_paused_resources_total` | Gauge | Currently paused resources |
| `finops_estimated_savings_usd` | Gauge | Projected monthly savings |
| `finops_policy_matches_total` | Counter | Policy evaluation matches |
| `finops_actions_taken_total` | Counter | Enforcement actions by type |
| `finops_reactivations_total` | Counter | User-initiated reactivations |
| `finops_false_positives_total` | Counter | Reverted within 1 hour |

### Grafana Dashboard

Import the provided dashboard from `deploy/grafana/dashboard.json`:

- Real-time savings projection
- Top idle namespaces
- Actions taken vs reverted
- Policy effectiveness

## Security

### RBAC Permissions

The controller requires minimal permissions:

- **Read**: pods, deployments, services (for cost correlation)
- **Write**: deployments (scale only), annotations
- **No**: delete permissions

See [config/rbac/](config/rbac/) for complete RBAC definitions.

### Network Policies

Restrict controller traffic:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: finops-enforcer
spec:
  podSelector:
    matchLabels:
      app: finops-enforcer
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: opencost
      ports:
        - protocol: TCP
          port: 9003
```

## Troubleshooting

### Common Issues

**Issue: No resources being paused**

- Check policy scope matches target namespaces
- Verify OpenCost is returning cost data
- Review dry-run mode setting

**Issue: False positives**

- Increase `idleWindow` duration
- Add namespace exclusions
- Adjust traffic thresholds

**Issue: Metrics not appearing**

- Verify Prometheus ServiceMonitor
- Check controller logs for errors
- Confirm metrics port accessibility

See [RUNBOOK.md](docs/RUNBOOK.md) for detailed troubleshooting.

## Roadmap

- [ ] Azure Cost Management integration
- [ ] AWS Cost Explorer integration
- [ ] Multi-cluster support
- [ ] Advanced scheduling policies
- [ ] Cost anomaly detection
- [ ] Self-service policy management UI

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Author

**Built and maintained by Rajesh Ramesh**

- GitHub: [@rramesh17993](https://github.com/rramesh17993)
- Portfolio: Production-grade Kubernetes controllers and cloud infrastructure automation

## Acknowledgments

- [OpenCost](https://www.opencost.io/) for real-time Kubernetes cost metrics
- CNCF for fostering cloud-native cost management practices

---

**Built with restraint, shipped with confidence.**

*This is practical infrastructure automation that respects the humans who have to live with it.*
