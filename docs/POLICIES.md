# Policy Reference Guide

## Overview

EnforcementPolicies are the core configuration mechanism for FinOps Enforcer. They define:

1. **Which resources** to evaluate (scope)
2. **What qualifies as idle** (conditions)
3. **What action to take** (actions)
4. **When and how** to enforce (enforcement)

## Policy Specification

### Complete Example

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: my-policy
  namespace: finops-system
spec:
  scope:
    namespaces:
      include:
        - dev-*
        - staging-*
      exclude:
        - prod
        - kube-system
    labels:
      match:
        team: platform
      exclude:
        critical: "true"
  
  conditions:
    idleWindow: 48h
    minHourlyCost: 2.0
    trafficThreshold:
      requestsPerMinute: 0
    utilizationThreshold:
      cpu: 5%
      memory: 10%
  
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
  
  enforcement:
    dryRun: false
    maxActionsPerRun: 10
    cooldownWindow: 1h
  
  schedule:
    timezone: America/Los_Angeles
    activeHours:
      - days: [Mon, Tue, Wed, Thu, Fri]
        hours: [9, 17]
```

## Field Reference

### spec.scope

Defines which resources the policy applies to.

#### spec.scope.namespaces

**Required**

```yaml
namespaces:
  include:
    - dev-*        # Wildcard patterns supported
    - staging-*
    - feature-*
  exclude:
    - prod         # Explicit exclusions
    - kube-system
```

- **include**: List of namespace patterns to include (supports `*` wildcard)
- **exclude**: List of namespace patterns to exclude (takes precedence)

#### spec.scope.labels

**Optional**

```yaml
labels:
  match:
    team: platform      # Must have this label
    env: non-prod
  exclude:
    critical: "true"    # Skip if this label present
```

- **match**: Resources must have ALL these labels
- **exclude**: Resources with ANY of these labels are skipped

### spec.conditions

Defines what qualifies as "idle".

#### spec.conditions.idleWindow

**Required**

Duration a resource must be idle before action is taken.

```yaml
idleWindow: 48h    # 48 hours
idleWindow: 2d     # 2 days (not yet supported)
```

Supported formats: `48h`, `72h`, etc.

#### spec.conditions.minHourlyCost

**Required**

Minimum hourly cost to care about (filters out low-cost resources).

```yaml
minHourlyCost: 2.0   # $2/hour minimum
```

**Why this matters**: Avoids spending enforcement overhead on resources costing pennies.

#### spec.conditions.trafficThreshold

**Optional**

Traffic-based idle detection.

```yaml
trafficThreshold:
  requestsPerMinute: 0   # Zero traffic = idle
```

Currently only supports `requestsPerMinute: 0` (future: configurable thresholds).

#### spec.conditions.utilizationThreshold

**Optional**

Resource utilization thresholds.

```yaml
utilizationThreshold:
  cpu: 5%       # <5% CPU usage
  memory: 10%   # <10% memory usage
```

**Note**: Implementation requires metrics integration (roadmap item).

### spec.actions

Defines what to do when conditions are met.

#### spec.actions.type

**Required**

```yaml
type: scaleToZero   # Only supported action
```

Currently only `scaleToZero` is implemented.

#### spec.actions.notify

**Required**

```yaml
notify: slack   # Send Slack notification
notify: none    # No notification
```

#### spec.actions.reactivationAllowed

**Required**

```yaml
reactivationAllowed: true   # Allow one-click reactivation
```

Must be `true` for Slack interactive buttons.

### spec.enforcement

Enforcement constraints and guardrails.

#### spec.enforcement.dryRun

**Optional** (default: `false`)

```yaml
dryRun: true   # Test mode - no actual enforcement
```

Use this to test policy configuration safely.

#### spec.enforcement.maxActionsPerRun

**Optional** (default: controller-level setting)

```yaml
maxActionsPerRun: 5   # Max 5 resources paused per reconciliation
```

Limits blast radius. Controller-level default is 10.

#### spec.enforcement.cooldownWindow

**Optional**

```yaml
cooldownWindow: 1h   # Minimum 1 hour between actions on same resource
```

Prevents flapping.

### spec.schedule

**Optional**

Defines when policy is active.

```yaml
schedule:
  timezone: America/Los_Angeles
  activeHours:
    - days: [Mon, Tue, Wed, Thu, Fri]
      hours: [9, 17]      # 9 AM to 5 PM
    - days: [Sat, Sun]
      hours: [0, 23]      # All day weekends
```

- **timezone**: IANA timezone (e.g., `America/New_York`, `Europe/London`)
- **activeHours**: List of time windows when policy runs

## Common Patterns

### Pattern 1: Aggressive Dev Environment Cleanup

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: dev-aggressive-gc
  namespace: finops-system
spec:
  scope:
    namespaces:
      include: [dev-*]
  conditions:
    idleWindow: 24h
    minHourlyCost: 1.0
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
  enforcement:
    maxActionsPerRun: 10
```

**Use case**: Fast-moving dev teams with short-lived environments.

### Pattern 2: Conservative Staging

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: staging-conservative
  namespace: finops-system
spec:
  scope:
    namespaces:
      include: [staging-*]
  conditions:
    idleWindow: 72h
    minHourlyCost: 5.0
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
  enforcement:
    maxActionsPerRun: 3
    cooldownWindow: 2h
```

**Use case**: Staging environments with longer lifecycles and higher sensitivity.

### Pattern 3: Weekend Shutdown

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
    minHourlyCost: 0.5
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

**Use case**: Non-prod environments unused on weekends.

### Pattern 4: High-Cost Only

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: expensive-only
  namespace: finops-system
spec:
  scope:
    namespaces:
      include: [dev-*]
  conditions:
    idleWindow: 24h
    minHourlyCost: 20.0  # Only target expensive resources
  actions:
    type: scaleToZero
    notify: slack
    reactivationAllowed: true
```

**Use case**: Focus enforcement on high-cost resources first.

## Exclusion Mechanisms

### Annotation-Based Exclusion

Add this annotation to any deployment to exclude it from ALL policies:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  annotations:
    finops.io/exclude: "true"
```

### Label-Based Exclusion

Use policy label filters:

```yaml
spec:
  scope:
    labels:
      exclude:
        critical: "true"
        always-on: "true"
```

Then label deployments:

```yaml
metadata:
  labels:
    critical: "true"
```

## Troubleshooting Policies

### Policy Not Matching Expected Resources

1. Check namespace scope:
   ```bash
   kubectl get deployment -A | grep <pattern>
   ```

2. Verify CRD is applied:
   ```bash
   kubectl get enforcementpolicies -n finops-system
   ```

3. Check policy status:
   ```bash
   kubectl describe enforcementpolicy <name> -n finops-system
   ```

### Too Many False Positives

Increase `idleWindow`:
```yaml
conditions:
  idleWindow: 72h  # More conservative
```

Or increase `minHourlyCost`:
```yaml
conditions:
  minHourlyCost: 5.0  # Only target expensive resources
```

### Not Enough Enforcement

Check:
- `maxActionsPerRun` isn't too low
- `cooldownWindow` isn't too long
- OpenCost is returning cost data

## Best Practices

### Start with Dry-Run

Always test new policies in dry-run mode first:

```yaml
enforcement:
  dryRun: true
```

Monitor Slack notifications and metrics for a few days before enabling.

### Use Multiple Policies

Better to have multiple focused policies than one complex policy:

```yaml
# Good: Separate policies by environment type
- dev-idle-gc (aggressive)
- staging-idle-gc (conservative)
- weekend-shutdown (scheduled)

# Avoid: One policy trying to do everything
```

### Set Appropriate Cooldowns

Prevent flapping:

```yaml
enforcement:
  cooldownWindow: 1h  # Minimum recommended
```

### Monitor Policy Effectiveness

Check status regularly:

```bash
kubectl get enforcementpolicies -n finops-system -o wide
```

Look at:
- `matchedResources`: How many resources match
- `actionsPerformed`: How many actions taken
- `estimatedSavings`: Projected monthly savings

## Security Considerations

### Namespace Restrictions

Always exclude critical namespaces:

```yaml
scope:
  namespaces:
    exclude:
      - kube-system
      - kube-public
      - kube-node-lease
      - prod
      - production
```

### Production Safeguards

**Never** include production namespaces without explicit approval:

```yaml
# DANGEROUS - requires explicit approval
scope:
  namespaces:
    include:
      - prod-*   # ⚠️ High risk
```

## Metrics

Policies expose metrics via Prometheus:

```promql
# Policy matches
finops_policy_matches_total{policy="dev-idle-gc"}

# Actions taken
finops_actions_taken_total{namespace="dev-payments", action="scaleToZero"}

# Estimated savings
finops_estimated_savings_usd{namespace="dev-payments"}
```

## See Also

- [DESIGN.md](DESIGN.md) - Architecture deep-dive
- [RUNBOOK.md](RUNBOOK.md) - Operational procedures
- [README.md](README.md) - Quick start guide
