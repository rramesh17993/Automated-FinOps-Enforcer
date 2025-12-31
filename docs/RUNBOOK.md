# Operational Runbook

## Overview

This runbook provides operational procedures for running FinOps Enforcer in production.

## Table of Contents

1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Monitoring](#monitoring)
4. [Common Operations](#common-operations)
5. [Troubleshooting](#troubleshooting)
6. [Emergency Procedures](#emergency-procedures)
7. [Maintenance](#maintenance)

---

## Installation

### Prerequisites

- Kubernetes cluster (1.24+)
- OpenCost installed and running
- Prometheus (for metrics)
- Slack workspace (optional, for notifications)

### Quick Install via Helm

```bash
# Add Helm repository (once published)
helm repo add finops-enforcer https://charts.finops-enforcer.io
helm repo update

# Install
helm install finops-enforcer finops-enforcer/finops-enforcer \
  --namespace finops-system \
  --create-namespace \
  --set opencost.endpoint=http://opencost.opencost:9003 \
  --set slack.webhookURL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Install via kubectl

```bash
# Clone repository
git clone https://github.com/yourusername/finops-enforcer.git
cd finops-enforcer

# Install CRDs
kubectl apply -f config/crd/

# Install controller
kubectl apply -f config/manager/namespace.yaml
kubectl apply -f config/rbac/
kubectl apply -f config/manager/

# Configure Slack (optional)
kubectl create secret generic finops-enforcer-secrets \
  --namespace finops-system \
  --from-literal=slack-webhook-url=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Verify Installation

```bash
# Check pod status
kubectl get pods -n finops-system

# Check logs
kubectl logs -n finops-system -l app=finops-enforcer -f

# Verify CRD
kubectl get crd enforcementpolicies.finops.io
```

---

## Configuration

### Controller Configuration

Edit the deployment:

```bash
kubectl edit deployment finops-enforcer -n finops-system
```

Key parameters:
- `--opencost-endpoint`: OpenCost API URL
- `--opencost-timeout`: API timeout (default: 30s)
- `--max-actions-per-run`: Global action limit (default: 10)
- `--leader-elect`: Enable for HA (default: false)

### OpenCost Integration

Verify OpenCost is accessible:

```bash
# Port-forward to OpenCost
kubectl port-forward -n opencost svc/opencost 9003:9003

# Test API
curl http://localhost:9003/allocation?window=2d
```

### Slack Integration

Create webhook in Slack:
1. Go to https://api.slack.com/apps
2. Create new app → Incoming Webhooks
3. Add to workspace
4. Copy webhook URL

Update secret:
```bash
kubectl create secret generic finops-enforcer-secrets \
  --namespace finops-system \
  --from-literal=slack-webhook-url=YOUR_WEBHOOK_URL \
  --dry-run=client -o yaml | kubectl apply -f -
```

---

## Monitoring

### Key Metrics

```promql
# Currently paused resources
finops_paused_resources_total

# Estimated monthly savings
sum(finops_estimated_savings_usd)

# Policy matches
rate(finops_policy_matches_total[5m])

# Actions taken
rate(finops_actions_taken_total[5m])

# False positive rate
rate(finops_false_positives_total[5m]) / rate(finops_actions_taken_total[5m])

# OpenCost errors
rate(finops_opencost_api_errors_total[5m])
```

### Alerts

Recommended Prometheus alerts:

```yaml
groups:
  - name: finops-enforcer
    rules:
      - alert: FinOpsEnforcerDown
        expr: up{job="finops-enforcer"} == 0
        for: 5m
        annotations:
          summary: "FinOps Enforcer is down"
      
      - alert: HighFalsePositiveRate
        expr: |
          rate(finops_false_positives_total[1h]) / 
          rate(finops_actions_taken_total[1h]) > 0.3
        for: 30m
        annotations:
          summary: "High false positive rate (>30%)"
      
      - alert: OpenCostAPIErrors
        expr: rate(finops_opencost_api_errors_total[5m]) > 0.1
        for: 10m
        annotations:
          summary: "Frequent OpenCost API errors"
```

### Grafana Dashboard

Import the dashboard from `deploy/grafana/dashboard.json`.

Key panels:
- Projected monthly savings (gauge)
- Resources paused over time (graph)
- Top idle namespaces (table)
- Policy effectiveness (heatmap)
- False positive rate (percentage)

---

## Common Operations

### Create a New Policy

```bash
# Start with dry-run
cat <<EOF | kubectl apply -f -
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: my-new-policy
  namespace: finops-system
spec:
  scope:
    namespaces:
      include: [test-*]
  conditions:
    idleWindow: 48h
    minHourlyCost: 2.0
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
  enforcement:
    dryRun: true  # Start safe
    maxActionsPerRun: 5
EOF

# Monitor for a few days
kubectl logs -n finops-system -l app=finops-enforcer | grep "DRY-RUN"

# Enable enforcement
kubectl patch enforcementpolicy my-new-policy -n finops-system \
  --type merge -p '{"spec":{"enforcement":{"dryRun":false}}}'
```

### List All Policies

```bash
kubectl get enforcementpolicies -n finops-system -o wide
```

### Check Policy Status

```bash
kubectl describe enforcementpolicy <name> -n finops-system
```

### View Paused Resources

```bash
# List all paused deployments
kubectl get deployments -A -l finops.io/paused=true

# See details
kubectl get deployment <name> -n <namespace> -o yaml | grep -A 10 "annotations:"
```

### Manually Reactivate a Resource

```bash
# Get original replica count
REPLICAS=$(kubectl get deployment <name> -n <namespace> \
  -o jsonpath='{.metadata.annotations.finops\.io/original-replicas}')

# Scale back up
kubectl scale deployment <name> -n <namespace> --replicas=$REPLICAS

# Clean up pause annotation
kubectl annotate deployment <name> -n <namespace> \
  finops.io/paused- \
  finops.io/paused-at-
```

### Exclude a Deployment from All Policies

```bash
kubectl annotate deployment <name> -n <namespace> \
  finops.io/exclude=true
```

### Update Controller Configuration

```bash
# Change max actions per run
kubectl set env deployment/finops-enforcer -n finops-system \
  MAX_ACTIONS_PER_RUN=20

# Update OpenCost endpoint
kubectl set env deployment/finops-enforcer -n finops-system \
  OPENCOST_ENDPOINT=http://new-opencost:9003
```

---

## Troubleshooting

### No Resources Being Paused

**Symptoms**: Policy exists but no enforcement actions

**Diagnosis**:
```bash
# Check policy status
kubectl get enforcementpolicies -n finops-system -o wide

# Check controller logs
kubectl logs -n finops-system -l app=finops-enforcer --tail=100

# Verify OpenCost
kubectl port-forward -n opencost svc/opencost 9003:9003
curl http://localhost:9003/allocation?window=2d
```

**Common causes**:
1. Dry-run mode enabled (`spec.enforcement.dryRun: true`)
2. Namespace scope doesn't match resources
3. Cost below `minHourlyCost` threshold
4. `idleWindow` not yet elapsed
5. OpenCost not returning data

**Resolution**:
```bash
# Disable dry-run
kubectl patch enforcementpolicy <name> -n finops-system \
  --type merge -p '{"spec":{"enforcement":{"dryRun":false}}}'

# Lower thresholds for testing
kubectl patch enforcementpolicy <name> -n finops-system \
  --type merge -p '{"spec":{"conditions":{"minHourlyCost":0.1,"idleWindow":"1h"}}}'
```

### High False Positive Rate

**Symptoms**: Resources frequently reactivated within 1 hour

**Diagnosis**:
```promql
rate(finops_false_positives_total[1h]) / rate(finops_actions_taken_total[1h])
```

**Resolution**:
```yaml
# Increase idle window
conditions:
  idleWindow: 72h  # Was 48h

# Increase cost threshold
conditions:
  minHourlyCost: 5.0  # Was 2.0

# Add cooldown
enforcement:
  cooldownWindow: 2h
```

### OpenCost Connection Errors

**Symptoms**: Logs show "failed to fetch cost data"

**Diagnosis**:
```bash
# Check OpenCost health
kubectl get pods -n opencost

# Test connectivity from controller pod
kubectl exec -n finops-system deployment/finops-enforcer -- \
  wget -O- http://opencost.opencost:9003/healthz
```

**Resolution**:
```bash
# Verify OpenCost service
kubectl get svc -n opencost

# Check network policy
kubectl get networkpolicy -n finops-system

# Restart controller
kubectl rollout restart deployment/finops-enforcer -n finops-system
```

### Controller Pod Crashes

**Symptoms**: CrashLoopBackOff

**Diagnosis**:
```bash
# Check logs
kubectl logs -n finops-system -l app=finops-enforcer --previous

# Check events
kubectl get events -n finops-system --sort-by='.lastTimestamp'
```

**Common causes**:
1. Invalid Slack webhook URL
2. RBAC permissions missing
3. Memory limits too low

**Resolution**:
```bash
# Check RBAC
kubectl auth can-i update deployments --as=system:serviceaccount:finops-system:finops-enforcer

# Increase memory
kubectl set resources deployment/finops-enforcer -n finops-system \
  --limits=memory=1Gi \
  --requests=memory=256Mi
```

### Metrics Not Appearing in Prometheus

**Diagnosis**:
```bash
# Check ServiceMonitor
kubectl get servicemonitor -n finops-system

# Port-forward and test
kubectl port-forward -n finops-system svc/finops-enforcer-metrics 8080:8080
curl http://localhost:8080/metrics
```

**Resolution**:
```bash
# Ensure Prometheus has RBAC to discover ServiceMonitors
kubectl get clusterrole prometheus

# Check Prometheus logs
kubectl logs -n monitoring -l app=prometheus
```

---

## Emergency Procedures

### Emergency: Stop All Enforcement

**Scenario**: Something went wrong, need to stop immediately

```bash
# Option 1: Pause all policies (dry-run)
kubectl get enforcementpolicies -n finops-system -o name | \
  xargs -I {} kubectl patch {} --type merge -p '{"spec":{"enforcement":{"dryRun":true}}}'

# Option 2: Scale down controller
kubectl scale deployment finops-enforcer -n finops-system --replicas=0

# Option 3: Delete all policies (drastic)
kubectl delete enforcementpolicies -n finops-system --all
```

### Emergency: Mass Reactivation

**Scenario**: Need to restore all paused resources immediately

```bash
# List all paused deployments
kubectl get deployments -A -o json | \
  jq -r '.items[] | select(.metadata.annotations."finops.io/paused" == "true") | 
  "\(.metadata.namespace) \(.metadata.name) \(.metadata.annotations."finops.io/original-replicas")"' | \
  while read namespace name replicas; do
    echo "Restoring $namespace/$name to $replicas replicas"
    kubectl scale deployment $name -n $namespace --replicas=$replicas
    kubectl annotate deployment $name -n $namespace finops.io/paused- finops.io/paused-at-
  done
```

### Emergency: Uninstall

```bash
# Via Helm
helm uninstall finops-enforcer -n finops-system

# Via kubectl
kubectl delete -f config/manager/
kubectl delete -f config/rbac/
kubectl delete -f config/crd/
kubectl delete namespace finops-system
```

---

## Maintenance

### Upgrade Controller

#### Via Helm

```bash
# Update chart repository
helm repo update

# Check new version
helm search repo finops-enforcer

# Upgrade
helm upgrade finops-enforcer finops-enforcer/finops-enforcer \
  --namespace finops-system \
  --reuse-values
```

#### Via kubectl

```bash
# Pull new image
docker pull finops-enforcer:v0.2.0

# Update deployment
kubectl set image deployment/finops-enforcer -n finops-system \
  manager=finops-enforcer:v0.2.0

# Watch rollout
kubectl rollout status deployment/finops-enforcer -n finops-system
```

### Backup Policies

```bash
# Export all policies
kubectl get enforcementpolicies -n finops-system -o yaml > policies-backup.yaml

# Restore
kubectl apply -f policies-backup.yaml
```

### Log Rotation

Logs are handled by Kubernetes. To retain audit trail:

```bash
# Stream logs to external system
kubectl logs -n finops-system -l app=finops-enforcer -f | \
  fluent-bit -c /etc/fluent-bit/fluent-bit.conf
```

### Health Checks

```bash
# Controller health
kubectl get pods -n finops-system

# Health endpoint
kubectl port-forward -n finops-system svc/finops-enforcer 8081:8081
curl http://localhost:8081/healthz
curl http://localhost:8081/readyz

# Metrics endpoint
curl http://localhost:8080/metrics
```

---

## Performance Tuning

### High Resource Count

If managing >1000 deployments:

```yaml
# Increase reconciliation interval
spec:
  enforcement:
    maxActionsPerRun: 20  # Process more per run

# Increase controller resources
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

### High Availability

```bash
# Enable leader election
helm upgrade finops-enforcer finops-enforcer/finops-enforcer \
  --set enforcement.leaderElection=true \
  --set replicaCount=3
```

---

## Support

### Debug Mode

```bash
# Enable debug logging
kubectl set env deployment/finops-enforcer -n finops-system \
  LOG_LEVEL=debug
```

### Collect Diagnostics

```bash
#!/bin/bash
# Save as collect-diagnostics.sh

mkdir -p finops-diagnostics

# Controller logs
kubectl logs -n finops-system -l app=finops-enforcer --tail=1000 > \
  finops-diagnostics/controller-logs.txt

# Policies
kubectl get enforcementpolicies -n finops-system -o yaml > \
  finops-diagnostics/policies.yaml

# Paused resources
kubectl get deployments -A -l finops.io/paused=true -o yaml > \
  finops-diagnostics/paused-resources.yaml

# Events
kubectl get events -n finops-system --sort-by='.lastTimestamp' > \
  finops-diagnostics/events.txt

# Metrics snapshot
kubectl port-forward -n finops-system svc/finops-enforcer-metrics 8080:8080 &
PF_PID=$!
sleep 2
curl http://localhost:8080/metrics > finops-diagnostics/metrics.txt
kill $PF_PID

tar czf finops-diagnostics.tar.gz finops-diagnostics/
echo "Diagnostics saved to finops-diagnostics.tar.gz"
```

---

## See Also

- [README.md](../README.md) - Overview and quick start
- [DESIGN.md](../DESIGN.md) - Architecture details
- [POLICIES.md](POLICIES.md) - Policy configuration reference
