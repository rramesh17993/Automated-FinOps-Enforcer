# Getting Started - Automated FinOps Enforcer

Welcome! This guide will get you up and running in 10 minutes.

## What You'll Build

By the end of this guide, you'll have:

- ✅ FinOps Enforcer running in your cluster
- ✅ OpenCost providing real-time cost data
- ✅ A test policy evaluating your dev namespaces
- ✅ Slack notifications (optional)
- ✅ Prometheus metrics

## Prerequisites

```bash
# Required
- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3

# Optional
- Slack workspace
- Prometheus
```

## Step-by-Step Installation

### 1. Install OpenCost (5 minutes)

OpenCost provides the cost data:

```bash
# Add Helm repo
helm repo add opencost https://opencost.github.io/opencost-helm-chart
helm repo update

# Install OpenCost
helm install opencost opencost/opencost \
  --namespace opencost \
  --create-namespace

# Verify
kubectl get pods -n opencost
# Wait until STATUS shows Running
```

### 2. Install FinOps Enforcer (3 minutes)

```bash
# Clone repository
git clone https://github.com/yourusername/finops-enforcer.git
cd finops-enforcer

# Install via Helm
helm install finops-enforcer deploy/helm/finops-enforcer \
  --namespace finops-system \
  --create-namespace \
  --set opencost.endpoint=http://opencost.opencost:9003

# Verify
kubectl get pods -n finops-system
# Should show 2 pods running (HA deployment)
```

**Success!** The controller is now running.

### 3. Create Your First Policy (2 minutes)

Start with a **dry-run** policy to see what it would do:

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
        - dev-*        # Your dev namespaces
        - test-*
      exclude:
        - prod         # Never touch production!
  
  conditions:
    idleWindow: 24h    # Idle for 24 hours
    minHourlyCost: 1.0 # Costs at least $1/hour
  
  actions:
    type: scaleToZero
    notify: none       # No Slack yet
    reactivationAllowed: true
  
  enforcement:
    dryRun: true       # SAFE MODE - no actual changes
    maxActionsPerRun: 5
EOF
```

### 4. Monitor What It Finds (Ongoing)

Watch the logs to see what the policy identifies:

```bash
# Stream logs
kubectl logs -n finops-system -l app=finops-enforcer -f

# Look for lines like:
# level=info msg="DRY-RUN: would execute action" 
#   deployment=my-app namespace=dev-test 
#   reason="Idle for 24h, cost $3.20/hr"
```

Check policy status:

```bash
kubectl get enforcementpolicies -n finops-system -o wide

# Output:
# NAME              MATCHED   ACTIONS   SAVINGS   AGE
# my-first-policy   7         0         0         5m
```

The `MATCHED` column shows resources that match your policy!

## What's Happening?

Every 5 minutes, the controller:

1. ✅ Lists all deployments in `dev-*` and `test-*` namespaces
2. ✅ Gets cost data from OpenCost for each
3. ✅ Checks if they've been idle for 24+ hours
4. ✅ Checks if hourly cost >= $1.00
5. ✅ **Logs what it WOULD do** (dry-run mode)

**No actual changes yet!** This is safe exploration.

## Next Steps

### Enable Enforcement (When Ready)

After monitoring dry-run for a few days and you're confident:

```bash
kubectl patch enforcementpolicy my-first-policy -n finops-system \
  --type merge -p '{"spec":{"enforcement":{"dryRun":false}}}'
```

Now it will **actually pause** idle resources!

### Add Slack Notifications

Get notified when resources are paused:

1. **Create Slack webhook**:
   - Go to https://api.slack.com/apps
   - Create app → Incoming Webhooks
   - Activate and copy webhook URL

2. **Update controller**:
   ```bash
   kubectl create secret generic finops-enforcer-secrets \
     --namespace finops-system \
     --from-literal=slack-webhook-url=https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     --dry-run=client -o yaml | kubectl apply -f -
   
   # Restart controller to pick up secret
   kubectl rollout restart deployment/finops-enforcer -n finops-system
   ```

3. **Update policy**:
   ```bash
   kubectl patch enforcementpolicy my-first-policy -n finops-system \
     --type merge -p '{"spec":{"actions":{"notify":"slack"}}}'
   ```

You'll now get Slack messages like:

```
🚨 Idle Resource Paused

Deployment: invoice-worker
Namespace: dev-payments
Idle Duration: 48 hours
Estimated Monthly Savings: $2,304

⏯️ Reactivate Now | 📄 View Policy
```

### View Metrics

If you have Prometheus:

```bash
kubectl port-forward -n finops-system svc/finops-enforcer-metrics 8080:8080

# In another terminal:
curl http://localhost:8080/metrics | grep finops
```

Key metrics:
- `finops_paused_resources_total` - How many paused right now
- `finops_estimated_savings_usd` - Projected monthly savings
- `finops_actions_taken_total` - Enforcement actions

### Apply More Policies

Try the sample policies:

```bash
# Weekend shutdown
kubectl apply -f config/samples/weekend-staging-shutdown.yaml

# High-cost aggressive cleanup
kubectl apply -f config/samples/expensive-idle-gc.yaml
```

## Common Operations

### Manually Reactivate a Resource

```bash
# Get original replica count
REPLICAS=$(kubectl get deployment <name> -n <namespace> \
  -o jsonpath='{.metadata.annotations.finops\.io/original-replicas}')

# Scale back up
kubectl scale deployment <name> -n <namespace> --replicas=$REPLICAS
```

### Exclude a Deployment Permanently

```bash
kubectl annotate deployment <name> -n <namespace> \
  finops.io/exclude=true
```

### List All Paused Resources

```bash
kubectl get deployments -A -l finops.io/paused=true
```

### Stop All Enforcement (Emergency)

```bash
# Option 1: Enable dry-run on all policies
kubectl get enforcementpolicies -n finops-system -o name | \
  xargs -I {} kubectl patch {} --type merge \
    -p '{"spec":{"enforcement":{"dryRun":true}}}'

# Option 2: Scale down controller
kubectl scale deployment finops-enforcer -n finops-system --replicas=0
```

## Troubleshooting

### "No resources being paused"

1. **Check dry-run is disabled**:
   ```bash
   kubectl get enforcementpolicy my-first-policy -n finops-system -o yaml | grep dryRun
   ```

2. **Verify OpenCost is working**:
   ```bash
   kubectl port-forward -n opencost svc/opencost 9003:9003
   curl http://localhost:9003/allocation?window=2d
   ```

3. **Lower thresholds for testing**:
   ```bash
   kubectl patch enforcementpolicy my-first-policy -n finops-system \
     --type merge -p '{"spec":{"conditions":{"minHourlyCost":0.1,"idleWindow":"1h"}}}'
   ```

### "Controller pod crashing"

```bash
# Check logs
kubectl logs -n finops-system -l app=finops-enforcer --previous

# Common issues:
# - Invalid Slack webhook → Remove secret and restart
# - RBAC issues → Check: kubectl auth can-i update deployments --as=system:serviceaccount:finops-system:finops-enforcer
```

### "Too many false positives"

Increase idle window:

```bash
kubectl patch enforcementpolicy my-first-policy -n finops-system \
  --type merge -p '{"spec":{"conditions":{"idleWindow":"72h"}}}'
```

## Learn More

- 📖 [Full Documentation](README.md)
- 🏗️ [Architecture Design](DESIGN.md)
- ⚙️ [Policy Configuration](docs/POLICIES.md)
- 🔧 [Operations Runbook](docs/RUNBOOK.md)
- 🤝 [Contributing Guide](CONTRIBUTING.md)

## Quick Reference

```bash
# View all policies
kubectl get enforcementpolicies -n finops-system

# View policy details
kubectl describe enforcementpolicy <name> -n finops-system

# Controller logs
kubectl logs -n finops-system -l app=finops-enforcer -f

# Metrics
kubectl port-forward -n finops-system svc/finops-enforcer-metrics 8080:8080

# Health check
curl http://localhost:8081/healthz
```

## Support

- 🐛 **Issues**: https://github.com/rramesh17993/finops-enforcer/issues
- 💬 **Discussions**: https://github.com/rramesh17993/finops-enforcer/discussions
- 📧 **Email**: rramesh17993@gmail.com

---

**You're all set!** Start with dry-run, monitor for a few days, then enable enforcement. Welcome to active cost governance! 🎉
