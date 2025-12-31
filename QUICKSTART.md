# Quick Start Guide

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3 (optional, for Helm installation)

## 5-Minute Quickstart

### Step 1: Install OpenCost

OpenCost is required for cost data:

```bash
helm repo add opencost https://opencost.github.io/opencost-helm-chart
helm install opencost opencost/opencost \
  --namespace opencost \
  --create-namespace
```

Verify OpenCost is running:

```bash
kubectl get pods -n opencost
```

### Step 2: Install FinOps Enforcer

#### Option A: Via Helm (Recommended)

```bash
helm install finops-enforcer deploy/helm/finops-enforcer \
  --namespace finops-system \
  --create-namespace \
  --set opencost.endpoint=http://opencost.opencost:9003 \
  --set slack.webhookURL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

#### Option B: Via kubectl

```bash
# Clone repository
git clone https://github.com/yourusername/finops-enforcer.git
cd finops-enforcer

# Install
kubectl apply -f config/manager/namespace.yaml
kubectl apply -f config/crd/
kubectl apply -f config/rbac/
kubectl apply -f config/manager/

# Configure Slack (optional)
kubectl create secret generic finops-enforcer-secrets \
  --namespace finops-system \
  --from-literal=slack-webhook-url=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Step 3: Verify Installation

```bash
# Check controller is running
kubectl get pods -n finops-system

# View logs
kubectl logs -n finops-system -l app=finops-enforcer -f
```

### Step 4: Create Your First Policy (Dry-Run)

Start safely with a test policy:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: my-first-policy
  namespace: finops-system
spec:
  scope:
    namespaces:
      include:
        - dev-*
      exclude:
        - prod
  
  conditions:
    idleWindow: 48h
    minHourlyCost: 2.0
  
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
  
  enforcement:
    dryRun: true  # Safe mode - no actual enforcement
    maxActionsPerRun: 5
EOF
```

### Step 5: Monitor Dry-Run Results

Watch what the policy *would* do:

```bash
# View policy status
kubectl get enforcementpolicies -n finops-system -o wide

# Check logs for DRY-RUN messages
kubectl logs -n finops-system -l app=finops-enforcer | grep "DRY-RUN"

# Check Slack notifications (if configured)
```

### Step 6: Enable Enforcement (When Ready)

After monitoring dry-run for a few days:

```bash
kubectl patch enforcementpolicy my-first-policy -n finops-system \
  --type merge -p '{"spec":{"enforcement":{"dryRun":false}}}'
```

## What Happens Next?

Every 5 minutes, the controller will:

1. Evaluate all deployments in `dev-*` namespaces
2. Check if they've been idle for 48+ hours
3. Verify hourly cost > $2
4. If matched:
   - Scale deployment to zero
   - Send Slack notification
   - Record estimated savings
   - Allow one-click reactivation

## Common Commands

```bash
# List all policies
kubectl get enforcementpolicies -n finops-system

# View policy details
kubectl describe enforcementpolicy my-first-policy -n finops-system

# List paused deployments
kubectl get deployments -A -l finops.io/paused=true

# View controller logs
kubectl logs -n finops-system -l app=finops-enforcer -f

# Manually reactivate a deployment
kubectl scale deployment <name> -n <namespace> --replicas=<original-count>
```

## Sample Use Cases

### Weekend Shutdown

Pause non-prod environments on weekends:

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: weekend-shutdown
  namespace: finops-system
spec:
  scope:
    namespaces:
      include: [dev-*, staging-*]
  conditions:
    idleWindow: 4h
    minHourlyCost: 1.0
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
  schedule:
    timezone: America/Los_Angeles
    activeHours:
      - days: [Sat, Sun]
        hours: [0, 23]
```

### High-Cost Idle Detection

Target expensive resources first:

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: expensive-idle-gc
  namespace: finops-system
spec:
  scope:
    namespaces:
      include: [dev-*]
  conditions:
    idleWindow: 24h
    minHourlyCost: 10.0  # Only expensive resources
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
  enforcement:
    maxActionsPerRun: 3  # Conservative
    cooldownWindow: 2h
```

## Excluding Resources

To exclude a deployment from all policies:

```bash
kubectl annotate deployment <name> -n <namespace> \
  finops.io/exclude=true
```

## Monitoring

### Metrics

Access Prometheus metrics:

```bash
kubectl port-forward -n finops-system svc/finops-enforcer-metrics 8080:8080
curl http://localhost:8080/metrics
```

Key metrics:
- `finops_paused_resources_total` - Currently paused
- `finops_estimated_savings_usd` - Projected monthly savings
- `finops_actions_taken_total` - Actions executed

### Grafana

Import dashboard from `deploy/grafana/dashboard.json`

## Troubleshooting

**No resources being paused?**

1. Check policy is not in dry-run mode
2. Verify namespace scope matches your resources
3. Check OpenCost is returning cost data
4. Lower `minHourlyCost` for testing

**Too many false positives?**

1. Increase `idleWindow` (e.g., 72h instead of 48h)
2. Increase `minHourlyCost` threshold
3. Add `cooldownWindow` to prevent flapping

**Controller not starting?**

1. Check RBAC permissions
2. Verify OpenCost endpoint is accessible
3. Check Slack webhook URL (if configured)

## Next Steps

- Read [POLICIES.md](docs/POLICIES.md) for configuration details
- Review [DESIGN.md](DESIGN.md) for architecture
- See [RUNBOOK.md](docs/RUNBOOK.md) for operations

## Support

- GitHub Issues: https://github.com/yourusername/finops-enforcer/issues
- Documentation: https://github.com/yourusername/finops-enforcer/docs

---

**Welcome to active cost governance!** 🎉
