# Design Document: Automated FinOps Enforcer

**RFC-001: Active Cost Governance for Kubernetes**

**Status:** Approved  
**Author:** Infrastructure Engineering  
**Last Updated:** 2025-12-31  
**Reviewers:** Platform, SRE, FinOps

---

## Executive Summary

Cloud cost overruns are rarely caused by bad intentions. They are caused by idle infrastructure that no one feels accountable for.

Most FinOps tooling today is:
- **Passive** (dashboards)
- **Retrospective** (after money is spent)
- **Advisory** (no enforcement)

This document proposes an **active, explainable, and reversible** cost governance system for Kubernetes that:

1. Detects idle or wasteful workloads in near real-time
2. Takes safe, bounded corrective action
3. Keeps humans in the loop
4. Produces measurable savings

**Core philosophy:** Dashboards don't save money. Actions do.

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Goals & Non-Goals](#goals--non-goals)
3. [Architecture](#architecture)
4. [Policy Model](#policy-model)
5. [Enforcement Logic](#enforcement-logic)
6. [Safety Guardrails](#safety-guardrails)
7. [Human Interaction](#human-interaction)
8. [Observability](#observability)
9. [Security Model](#security-model)
10. [Failure Modes](#failure-modes)
11. [Trade-Offs](#trade-offs)
12. [Success Metrics](#success-metrics)
13. [What Was NOT Built](#what-was-not-built)

---

## Problem Statement

### The Reality on the Ground

In most organizations with Kubernetes:

- Dev and staging environments run 24×7
- Teams forget to tear down experiments
- Non-prod workloads quietly accumulate
- Bills spike, and no one knows who or why

By the time Cost Explorer or FinOps dashboards are checked:
- The money is already gone
- The signal is noisy
- Accountability is diffused

### Why Existing Tools Fall Short

| Tool Category | Limitation |
|--------------|------------|
| Cloud Cost Dashboards | Passive, retrospective |
| Alerts on Spend | Fire too late |
| Budget Caps | Blunt, often bypassed |
| Manual Cleanup | Doesn't scale |

### Concrete Impact

**Before this system:**
- Non-prod infrastructure cost: $18,000/month
- Idle resource percentage: ~40%
- Wasted spend: $7,200/month
- Manual cleanup overhead: 8 hours/week

**After this system:**
- Automatic idle detection and pausing
- Projected savings: $5,000+/month
- Manual overhead: <1 hour/week (policy tuning)
- False positive rate: <5%

---

## Goals & Non-Goals

### Goals

The system MUST:

1. **Detect idle Kubernetes workloads** with high confidence
2. **Act automatically** but conservatively
3. Be **explainable** (every action has a reason)
4. Be **reversible** (nothing destructive)
5. **Produce clear savings metrics**
6. **Preserve developer trust** (no surprise deletions)

### Non-Goals

This section is critical for scope control and interview discussions.

❌ **No ML cost forecasting**  
   - Reason: Adds complexity, low ROI for initial version
   - Decision: Use simple heuristics first

❌ **No billing system replacement**  
   - Reason: Out of scope, already solved
   - Decision: Integrate with existing tools

❌ **No organization-wide global policies**  
   - Reason: Too risky, insufficient context
   - Decision: Namespace-scoped, opt-in model

❌ **No hard deletion of resources**  
   - Reason: Trust and safety
   - Decision: Scale-to-zero only

❌ **No enforcement in production namespaces (by default)**  
   - Reason: Too high risk
   - Decision: Explicit opt-in required

**Philosophy:** This project is about practical governance, not theoretical optimization.

---

## Architecture

### High-Level Components

```
┌─────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                    │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ Deployments  │  │  Services    │  │   Pods       │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
          ┌───────────────────────────────┐
          │         OpenCost              │
          │  (Real-time Cost Metrics)     │
          │                               │
          │  - Cost per namespace         │
          │  - Cost per deployment        │
          │  - Cost by label              │
          └───────────────┬───────────────┘
                          │
                          ▼
          ┌────────────────────────────────────┐
          │   FinOps Enforcer Controller       │
          │                                    │
          │  ┌──────────────────────────────┐ │
          │  │    Policy Engine             │ │
          │  │  - Load policies (CRDs)      │ │
          │  │  - Evaluate conditions       │ │
          │  └──────────────────────────────┘ │
          │                                    │
          │  ┌──────────────────────────────┐ │
          │  │    Cost Analyzer             │ │
          │  │  - Fetch OpenCost data       │ │
          │  │  - Correlate with workloads  │ │
          │  └──────────────────────────────┘ │
          │                                    │
          │  ┌──────────────────────────────┐ │
          │  │    Enforcement Engine        │ │
          │  │  - Scale deployments         │ │
          │  │  - Annotate resources        │ │
          │  │  - Rate limiting             │ │
          │  └──────────────────────────────┘ │
          │                                    │
          │  ┌──────────────────────────────┐ │
          │  │    Notification Manager      │ │
          │  │  - Slack integration         │ │
          │  │  - Reactivation handler      │ │
          │  └──────────────────────────────┘ │
          │                                    │
          │  ┌──────────────────────────────┐ │
          │  │    Metrics Exporter          │ │
          │  │  - Prometheus metrics        │ │
          │  │  - Savings calculations      │ │
          │  └──────────────────────────────┘ │
          └────────────────────────────────────┘
                          │
                          ▼
          ┌────────────────────────────────────┐
          │      Observability Stack           │
          │                                    │
          │  ┌──────────────┐ ┌─────────────┐ │
          │  │ Prometheus   │ │   Grafana   │ │
          │  └──────────────┘ └─────────────┘ │
          │                                    │
          │  ┌──────────────┐ ┌─────────────┐ │
          │  │    Slack     │ │ Audit Logs  │ │
          │  └──────────────┘ └─────────────┘ │
          └────────────────────────────────────┘
```

### Component Responsibilities

#### 1. Cost Visibility Layer (OpenCost)

**What:** CNCF project providing real-time Kubernetes cost allocation.

**Why:** Need accurate, per-resource cost data without polling cloud APIs.

**Integration:**
- Deployed via Helm
- Exposes Prometheus metrics
- REST API for querying costs by namespace/label/deployment

#### 2. Policy Engine

**What:** Declarative rule system defining "idle" conditions.

**Why:** Business logic must be auditable, versioned, and non-developer-readable.

**Implementation:**
- Custom Resource Definitions (CRDs)
- Stored in Kubernetes etcd
- Hot-reloadable without controller restart

#### 3. Enforcement Controller

**What:** Kubernetes controller reconciling policy against reality.

**Why:** Native Kubernetes patterns for reliability.

**Implementation:**
- Built with controller-runtime
- Reconciliation loop every 5 minutes (configurable)
- Leader election for HA

#### 4. Human Feedback Loop

**What:** Slack notifications with interactive buttons.

**Why:** Preserve trust, enable quick reactivation.

**Implementation:**
- Slack webhooks for outbound
- Slack Slash commands for reactivation
- Annotate paused resources with reactivation instructions

#### 5. Reporting & Audit

**What:** Prometheus metrics + Grafana dashboards.

**Why:** Measure impact, justify continued operation.

**Metrics:**
- `finops_paused_resources_total`
- `finops_estimated_savings_usd`
- `finops_policy_matches_total`
- `finops_actions_taken_total`

---

## Policy Model

### Design Principles

1. **Config-as-Code:** Policies are YAML, versioned in Git
2. **Auditable:** Every change goes through PR
3. **Scoped:** Policies apply to specific namespaces only
4. **Composable:** Multiple policies can coexist
5. **Safe Defaults:** Dry-run mode by default

### Policy CRD Structure

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: dev-idle-gc
  namespace: finops-system
spec:
  # Define which resources this policy applies to
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

  # Define what qualifies as "idle"
  conditions:
    # Resource must be idle for this duration
    idleWindow: 48h
    
    # Minimum hourly cost to care about
    minHourlyCost: 2.0
    
    # Traffic thresholds
    trafficThreshold:
      requestsPerMinute: 0
      
    # CPU/Memory utilization (optional)
    utilizationThreshold:
      cpu: 5%
      memory: 10%

  # Define what action to take
  actions:
    type: scaleToZero  # Only supported action
    notify: slack
    reactivationAllowed: true
    
  # Enforcement constraints
  enforcement:
    dryRun: false
    maxActionsPerRun: 5
    cooldownWindow: 1h

  # Schedule when policy is active
  schedule:
    timezone: America/Los_Angeles
    activeHours:
      - days: [Mon, Tue, Wed, Thu, Fri]
        hours: [9, 17]  # 9 AM to 5 PM
```

### Example Policies

#### Policy 1: Weekend Non-Prod Cleanup

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: weekend-idle-gc
spec:
  scope:
    namespaces:
      include: [dev-*, staging-*]
  conditions:
    idleWindow: 4h
    minHourlyCost: 1.0
  schedule:
    activeHours:
      - days: [Sat, Sun]
        hours: [0, 23]
```

#### Policy 2: High-Cost Idle Detection

```yaml
apiVersion: finops.io/v1alpha1
kind: EnforcementPolicy
metadata:
  name: expensive-idle-gc
spec:
  scope:
    namespaces:
      include: [dev-*]
  conditions:
    idleWindow: 24h
    minHourlyCost: 10.0  # Flag expensive resources faster
  actions:
    type: scaleToZero
    notify: slack
```

---

## Enforcement Logic

### Reconciliation Loop

Every 5 minutes (configurable), the controller:

1. **Fetch all EnforcementPolicies** from Kubernetes API
2. **For each policy:**
   - Query OpenCost for resources in scope
   - Correlate cost data with Kubernetes resources
   - Evaluate idle conditions
   - Execute actions if criteria met
3. **Rate limit:** Max 10 actions per run
4. **Audit:** Log all decisions (match/no-match/action-taken)

### Decision Tree

```
For each Deployment in scope:
  ├─ Is namespace in policy scope?
  │  ├─ NO → Skip
  │  └─ YES → Continue
  │
  ├─ Does it have finops/exclude=true annotation?
  │  ├─ YES → Skip
  │  └─ NO → Continue
  │
  ├─ Is it already paused (replicas=0, finops/paused=true)?
  │  ├─ YES → Skip
  │  └─ NO → Continue
  │
  ├─ Is hourly cost >= minHourlyCost?
  │  ├─ NO → Skip
  │  └─ YES → Continue
  │
  ├─ Has it been idle for >= idleWindow?
  │  ├─ NO → Skip
  │  └─ YES → Continue
  │
  ├─ Is traffic below threshold?
  │  ├─ NO → Skip
  │  └─ YES → Continue
  │
  ├─ Is cooldown window expired?
  │  ├─ NO → Skip
  │  └─ YES → Execute action
  │
  └─ ACTION:
     ├─ Store original replica count in annotation
     ├─ Scale deployment to zero
     ├─ Add finops/paused=true annotation
     ├─ Add finops/paused-at=<timestamp> annotation
     ├─ Send Slack notification
     └─ Increment metrics
```

### Action Execution

```go
// Pseudo-code for clarity
func enforcePolicy(deployment *appsv1.Deployment, policy *finopsv1alpha1.EnforcementPolicy) error {
    // Store original state
    originalReplicas := *deployment.Spec.Replicas
    
    // Create action record
    action := &Action{
        Type: "scaleToZero",
        Resource: deployment.Name,
        Namespace: deployment.Namespace,
        Timestamp: time.Now(),
        Reason: "Idle for 48h, zero traffic, cost $3.20/hr",
        EstimatedMonthlySavings: calculateSavings(deployment),
    }
    
    // Annotate deployment
    if deployment.Annotations == nil {
        deployment.Annotations = make(map[string]string)
    }
    deployment.Annotations["finops.io/paused"] = "true"
    deployment.Annotations["finops.io/paused-at"] = action.Timestamp.Format(time.RFC3339)
    deployment.Annotations["finops.io/original-replicas"] = fmt.Sprintf("%d", originalReplicas)
    deployment.Annotations["finops.io/policy"] = policy.Name
    deployment.Annotations["finops.io/reason"] = action.Reason
    
    // Scale to zero
    replicas := int32(0)
    deployment.Spec.Replicas = &replicas
    
    // Update in Kubernetes
    if err := r.Client.Update(ctx, deployment); err != nil {
        return fmt.Errorf("failed to pause deployment: %w", err)
    }
    
    // Send notification
    r.NotificationManager.SendSlackAlert(action)
    
    // Update metrics
    r.Metrics.PausedResourcesTotal.Inc()
    r.Metrics.EstimatedSavingsUSD.Add(action.EstimatedMonthlySavings)
    
    return nil
}
```

---

## Safety Guardrails

### Principle: Fail Closed

If anything is uncertain, do nothing.

### Guardrails Implemented

| Guardrail | Purpose | Implementation |
|-----------|---------|----------------|
| **Namespace Allowlist** | Prevent prod accidents | Explicit include/exclude lists |
| **Cooldown Window** | Prevent flapping | 1-hour minimum between actions on same resource |
| **Max Actions Per Run** | Limit blast radius | Default: 10 resources/reconciliation |
| **Dry-Run Mode** | Safe testing | Policy-level flag, no actual scaling |
| **Annotation Exclusions** | Manual overrides | `finops.io/exclude=true` skips resource |
| **Minimum Cost Threshold** | Avoid noise | Only act on resources costing >$2/hour |
| **Reactivation Path** | Easy rollback | One-click Slack button |

### Dry-Run Mode

**Purpose:** Test policies without actual enforcement.

**Behavior:**
- Evaluate all conditions
- Log what WOULD happen
- Emit metrics with `dry_run=true` label
- Send Slack notification marked as DRY-RUN

**Usage:**
```yaml
spec:
  enforcement:
    dryRun: true  # Safe mode
```

### Cooldown Implementation

```go
func isCooldownExpired(deployment *appsv1.Deployment, cooldownWindow time.Duration) bool {
    pausedAtStr := deployment.Annotations["finops.io/paused-at"]
    if pausedAtStr == "" {
        return true  // Never paused before
    }
    
    pausedAt, err := time.Parse(time.RFC3339, pausedAtStr)
    if err != nil {
        return true  // Can't parse, assume expired
    }
    
    return time.Since(pausedAt) >= cooldownWindow
}
```

---

## Human Interaction

### Slack Notification Design

**Principles:**
- Clear, actionable information
- One-click reactivation
- No jargon

**Example Message:**

```
🚨 Idle Resource Paused

Namespace: dev-payments
Deployment: invoice-worker
Idle Duration: 72 hours
Last Traffic: 2025-12-28 14:32 UTC
Hourly Cost: $3.20
Estimated Monthly Savings: $2,304

Reason: Zero requests for 3 days, CPU utilization <5%

⏯️ Reactivate Now | 📄 View Policy | 🔍 View Logs

To reactivate manually:
kubectl scale deployment invoice-worker -n dev-payments --replicas=3
```

### Reactivation Flow

**User clicks "Reactivate Now" button:**

1. Slack sends callback to controller webhook
2. Controller validates request
3. Reads `finops.io/original-replicas` annotation
4. Scales deployment back to original count
5. Removes `finops.io/paused` annotation
6. Sends confirmation to Slack
7. Updates metrics (`finops_reactivations_total`)

**Implementation:**

```go
func handleReactivation(w http.ResponseWriter, r *http.Request) {
    var payload SlackPayload
    json.NewDecoder(r.Body).Decode(&payload)
    
    namespace := payload.Namespace
    deploymentName := payload.Deployment
    
    deployment := &appsv1.Deployment{}
    err := k8sClient.Get(ctx, types.NamespacedName{
        Name: deploymentName,
        Namespace: namespace,
    }, deployment)
    
    if err != nil {
        respondError(w, "Deployment not found")
        return
    }
    
    // Read original replica count
    originalReplicasStr := deployment.Annotations["finops.io/original-replicas"]
    originalReplicas, _ := strconv.ParseInt(originalReplicasStr, 10, 32)
    
    // Restore
    replicas := int32(originalReplicas)
    deployment.Spec.Replicas = &replicas
    
    // Clean up annotations
    delete(deployment.Annotations, "finops.io/paused")
    delete(deployment.Annotations, "finops.io/paused-at")
    
    k8sClient.Update(ctx, deployment)
    
    // Respond to Slack
    respondSuccess(w, fmt.Sprintf("✅ Reactivated %s in %s (scaled to %d replicas)", 
        deploymentName, namespace, originalReplicas))
    
    // Metrics
    reactivationsTotal.Inc()
}
```

---

## Observability

### Metrics Exposed

All metrics follow Prometheus naming conventions and include relevant labels.

#### Core Metrics

```
# HELP finops_paused_resources_total Number of resources currently paused by FinOps Enforcer
# TYPE finops_paused_resources_total gauge
finops_paused_resources_total{namespace="dev-payments",policy="dev-idle-gc"} 3

# HELP finops_estimated_savings_usd Estimated monthly savings from paused resources
# TYPE finops_estimated_savings_usd gauge
finops_estimated_savings_usd{namespace="dev-payments"} 2304.50

# HELP finops_policy_matches_total Number of times policies matched resources
# TYPE finops_policy_matches_total counter
finops_policy_matches_total{policy="dev-idle-gc",action="scaleToZero"} 47

# HELP finops_actions_taken_total Number of enforcement actions taken
# TYPE finops_actions_taken_total counter
finops_actions_taken_total{action="scaleToZero",namespace="dev-payments"} 15

# HELP finops_reactivations_total Number of user-initiated reactivations
# TYPE finops_reactivations_total counter
finops_reactivations_total{namespace="dev-payments",source="slack"} 4

# HELP finops_false_positives_total Resources reactivated within 1 hour (likely false positive)
# TYPE finops_false_positives_total counter
finops_false_positives_total{namespace="dev-payments"} 1

# HELP finops_policy_evaluation_duration_seconds Time spent evaluating policies
# TYPE finops_policy_evaluation_duration_seconds histogram
finops_policy_evaluation_duration_seconds_bucket{le="0.1"} 42
```

### Grafana Dashboard

**Panels:**

1. **Projected Monthly Savings** (gauge)
2. **Resources Paused** (time series)
3. **Top Idle Namespaces** (bar chart)
4. **Actions Taken vs Reverted** (dual axis)
5. **Policy Match Rate** (heatmap)
6. **False Positive Rate** (percentage)

**Focus:** Impact metrics, not vanity metrics.

### Audit Logging

Every action is logged to stdout in structured JSON:

```json
{
  "timestamp": "2025-12-31T10:15:32Z",
  "level": "info",
  "msg": "enforcement_action_taken",
  "action": "scaleToZero",
  "namespace": "dev-payments",
  "deployment": "invoice-worker",
  "policy": "dev-idle-gc",
  "reason": "Idle for 72h, zero traffic",
  "original_replicas": 3,
  "estimated_monthly_savings_usd": 2304.50,
  "dry_run": false
}
```

---

## Security Model

### RBAC Permissions

The controller service account requires minimal permissions:

**Read:**
- `pods` (for usage metrics)
- `deployments` (for evaluation)
- `services` (for traffic correlation)
- `enforcementpolicies` (CRD)

**Write:**
- `deployments` (scale operation only)
- `deployments` (annotations only)

**Forbidden:**
- `delete` on any resource
- Write to `kube-system` or `finops-system`

**RBAC Manifest:**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: finops-enforcer-role
rules:
  # Read-only for cost correlation
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch"]
  
  - apiGroups: [""]
    resources: ["pods", "services"]
    verbs: ["get", "list", "watch"]
  
  # Write for enforcement (scale + annotate only)
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["update", "patch"]
  
  # CRD access
  - apiGroups: ["finops.io"]
    resources: ["enforcementpolicies"]
    verbs: ["get", "list", "watch"]
  
  # Explicitly NO delete permissions
```

### Network Policies

Restrict controller network access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: finops-enforcer-netpol
  namespace: finops-system
spec:
  podSelector:
    matchLabels:
      app: finops-enforcer
  policyTypes:
    - Egress
  egress:
    # Allow Kubernetes API access
    - to:
        - namespaceSelector: {}
      ports:
        - protocol: TCP
          port: 443
    
    # Allow OpenCost access
    - to:
        - namespaceSelector:
            matchLabels:
              name: opencost
      ports:
        - protocol: TCP
          port: 9003
    
    # Allow Slack webhook
    - to:
        - podSelector: {}
      ports:
        - protocol: TCP
          port: 443
```

### Secrets Management

**Slack Webhook:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: finops-enforcer-secrets
  namespace: finops-system
type: Opaque
stringData:
  slack-webhook-url: https://hooks.slack.com/services/XXX/YYY/ZZZ
```

**Mounted as environment variable:**
```yaml
env:
  - name: SLACK_WEBHOOK_URL
    valueFrom:
      secretKeyRef:
        name: finops-enforcer-secrets
        key: slack-webhook-url
```

---

## Failure Modes

### Failure Mode Analysis

| Failure Scenario | Impact | Mitigation |
|------------------|--------|------------|
| OpenCost unavailable | No cost data | Fail closed, log error, retry |
| Kubernetes API throttling | Enforcement delayed | Backoff, rate limiting |
| Slack webhook down | No notifications | Queue messages, retry with exponential backoff |
| Policy misconfiguration | Incorrect enforcement | Dry-run mode, validation webhook |
| Controller crash | No enforcement until restart | Leader election, multiple replicas |
| Network partition | Split-brain risk | Leader election prevents dual writes |

### Error Handling Strategy

**Principle:** Fail closed, log clearly, recover gracefully.

**Implementation:**

```go
func reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    // Fetch cost data
    costData, err := r.CostClient.GetCosts(ctx, req.Namespace)
    if err != nil {
        // FAIL CLOSED: Do not enforce without cost data
        r.Log.Error(err, "failed to fetch cost data, skipping enforcement",
            "namespace", req.Namespace)
        return reconcile.Result{RequeueAfter: 5 * time.Minute}, nil
    }
    
    // Evaluate policies
    actions, err := r.PolicyEngine.Evaluate(ctx, costData)
    if err != nil {
        r.Log.Error(err, "policy evaluation failed, skipping enforcement")
        return reconcile.Result{RequeueAfter: 5 * time.Minute}, nil
    }
    
    // Execute actions
    for _, action := range actions {
        if err := r.Enforcer.Execute(ctx, action); err != nil {
            // LOG but continue (don't block other actions)
            r.Log.Error(err, "action execution failed",
                "action", action.Type,
                "resource", action.Resource)
            continue
        }
    }
    
    return reconcile.Result{RequeueAfter: 5 * time.Minute}, nil
}
```

---

## Trade-Offs

Explicit recognition of what was gained and what was accepted.

### What We Gain

✅ **Immediate, measurable savings** ($5,000+/month typical)  
✅ **Low operational overhead** (1 hour/week)  
✅ **Trustworthy automation** (reversible, explainable)  
✅ **Developer experience** (transparent, non-disruptive)  
✅ **Auditability** (every action logged)

### What We Accept

⚠️ **Some false positives** (<5% acceptable)  
⚠️ **Conservative scope** (non-prod only initially)  
⚠️ **No global optimization** (namespace-scoped)  
⚠️ **Eventual consistency** (5-minute reconciliation loop)  
⚠️ **Manual policy tuning** (not fully autonomous)

**This is by design.** Trust > absolute efficiency.

### Why These Trade-Offs Make Sense

**False Positives:**
- Easy reactivation (1 click)
- Better than false negatives (invisible cost bleed)

**Conservative Scope:**
- Production safety non-negotiable
- Expand after trust is earned

**No Global Optimization:**
- Reduces blast radius
- Teams retain autonomy

**Eventual Consistency:**
- 5-minute delay acceptable for cost management
- Reduces API load

---

## Success Metrics

How we measure if this system is succeeding.

### Primary Metrics

1. **$ Reduction in non-prod spend** (monthly)
   - Target: 30% reduction
   - Baseline: $18,000/month
   - Goal: $12,000/month

2. **Mean Time to Idle Detection**
   - Current: N/A (manual)
   - Target: <6 hours

3. **Number of Safe Auto-Pauses**
   - Target: >20/week
   - Must maintain <5% false positive rate

4. **Human Override Rate** (trust signal)
   - <10% reactivation within 1 hour = good signal quality
   - >30% reactivation = policy too aggressive

### Secondary Metrics

5. **Policy Coverage**
   - % of non-prod namespaces with policies
   - Target: >80%

6. **Time to Reactivation**
   - Target: <2 minutes from Slack notification to restored

7. **Developer NPS**
   - Survey: "How do you feel about FinOps Enforcer?"
   - Target: >0 (net positive)

---

## What Was NOT Built

This section is critical for demonstrating judgment and restraint.

### Explicitly Excluded

❌ **No ML cost forecasting**
- **Why:** Adds complexity without proven ROI
- **Instead:** Use simple heuristics (idle window, traffic threshold)
- **Future:** May add if heuristics prove insufficient

❌ **No cloud billing ingestion**
- **Why:** OpenCost is sufficient for Kubernetes
- **Instead:** Integrate with existing cloud billing tools
- **Future:** Consider if multi-cloud scope expands

❌ **No organization-wide global policies**
- **Why:** Too risky, insufficient context
- **Instead:** Namespace-scoped, opt-in model
- **Future:** May add org-level defaults with namespace overrides

❌ **No cross-account orchestration**
- **Why:** Out of scope, adds auth complexity
- **Instead:** Single-cluster focus initially
- **Future:** Multi-cluster with careful design

❌ **No attempt to replace FinOps teams**
- **Why:** This is an automation tool, not a strategy replacement
- **Instead:** Provide actionable data for FinOps teams to use
- **Future:** Better integration with FinOps workflows

### Why This Matters

**Interview perspective:**

> "We could have built ML forecasting models, but that would have delayed delivery by 3 months for uncertain value. Instead, we shipped a working system in 6 weeks that saves real money today. If heuristics prove insufficient, we have telemetry to justify ML investment."

This demonstrates:
- Pragmatism
- Incremental delivery
- Data-driven decision making
- Understanding of opportunity cost

---

## Appendix

### Glossary

**Idle Workload:** Resource with non-zero cost but zero value signals (traffic, CPU, etc.)

**Zombie Resource:** Another term for idle workload (deprecated internally)

**Scale-to-Zero:** Setting `deployment.spec.replicas = 0`

**Cooldown Window:** Minimum time between enforcement actions on same resource

**Dry-Run Mode:** Policy evaluation without actual enforcement

**False Positive:** Resource paused incorrectly, reactivated within 1 hour

### References

- [OpenCost Documentation](https://www.opencost.io/)
- [Kubernetes Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [FinOps Foundation](https://www.finops.org/)
- [Prometheus Metrics Best Practices](https://prometheus.io/docs/practices/naming/)

### Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-12-31 | 1.0 | Initial design document |

---

**Approved by:** Infrastructure Leadership  
**Implementation Status:** In Progress  
**Target GA:** Q1 2026
