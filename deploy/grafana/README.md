# Grafana Dashboards for FinOps Enforcer

This directory contains pre-built Grafana dashboards for visualizing business and operational metrics.

## Business Metrics Dashboard

**File:** `business-metrics-dashboard.json`

### Key Panels

1. **💰 Monthly Savings Trajectory**
   - Large stat showing projected monthly savings
   - Color-coded: Red (<$500), Yellow ($500-$2000), Green (>$2000)
   - Shows the money NOT being spent right now

2. **📈 ROI Calculator**
   - Annual ROI multiplier = (Annual Savings) / (Controller Overhead)
   - Assumes $200/month overhead (controller resources + operator time)
   - Example: $4,500/month = $54,000/year ÷ $2,400 = 22.5x ROI

3. **🎯 Enforcement Accuracy**
   - Percentage of actions that were correct (not reversed within 24h)
   - Formula: `(1 - false_positives / total_actions) * 100`
   - Target: >95% accuracy

4. **⚡ Resources Currently Paused**
   - Real-time count of deployments scaled to zero
   - Trend line shows historical pause activity

5. **📊 Actions Per Day**
   - Number of enforcement actions in last 24h
   - Helps identify busy periods

6. **💵 Savings Over Time**
   - Historical trend of monthly and annualized savings
   - Shows impact of policy changes over time

7. **🏆 Top Cost-Saving Policies**
   - Table ranking policies by savings generated
   - Helps prioritize which policies to tune/expand

8. **🔍 Resource Matches vs Actions**
   - Gap shows impact of rate limiting and dry-run mode
   - If gap is large, consider disabling dry-run

9. **⚠️ Policy Violations & Overrides**
   - Count of manually excluded resources
   - High number might indicate policy is too aggressive

10. **🔄 Reactivation Rate**
    - How many resources reactivated in last 24h
    - High rate = possible false positives

11. **⏱️ Average Reconciliation Time**
    - Controller performance metric
    - Alert if >30 seconds (might need optimization)

## Installation

### Option 1: Import via Grafana UI

1. Open Grafana
2. Navigate to **Dashboards** → **Import**
3. Upload `business-metrics-dashboard.json`
4. Select Prometheus data source
5. Click **Import**

### Option 2: Automated Deployment (ConfigMap)

```bash
# Create ConfigMap with dashboard
kubectl create configmap finops-grafana-dashboards \
  --from-file=business-metrics.json=business-metrics-dashboard.json \
  -n monitoring

# Annotate for Grafana sidecar discovery
kubectl annotate configmap finops-grafana-dashboards \
  grafana_dashboard="1" \
  -n monitoring
```

**Note:** Requires Grafana with sidecar enabled:
```yaml
sidecar:
  dashboards:
    enabled: true
    label: grafana_dashboard
```

### Option 3: Provision via Grafana Helm Chart

```yaml
# values.yaml for Grafana Helm chart
dashboardProviders:
  dashboardproviders.yaml:
    apiVersion: 1
    providers:
      - name: 'finops'
        orgId: 1
        folder: 'FinOps'
        type: file
        disableDeletion: false
        editable: true
        options:
          path: /var/lib/grafana/dashboards/finops

dashboards:
  finops:
    business-metrics:
      file: dashboards/business-metrics-dashboard.json
```

## Required Metrics

The dashboard expects these Prometheus metrics from the controller:

```promql
# Paused resources
finops_paused_resources_total{namespace="..."}

# Estimated savings
finops_estimated_savings_usd

# Actions taken
finops_actions_taken_total{policy="..."}

# Policy matches
finops_policy_matches_total{policy="..."}

# Reconciliation performance
finops_reconciliation_duration_seconds

# Reactivations
finops_reactivations_total

# Exclusions
finops_excluded_resources_total

# False positives (optional - requires tracking)
finops_false_positives_total
```

**Note:** If `finops_false_positives_total` is not implemented, the "Enforcement Accuracy" panel will show no data. To implement:

```go
// In pkg/metrics/metrics.go
var FalsePositivesTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "finops_false_positives_total",
        Help: "Total false positive enforcement actions (reactivated within 24h)",
    },
    []string{"namespace", "deployment", "policy"},
)

// Track when resources are reactivated within 24h of being paused
```

## Dashboard Variables

The dashboard includes template variables for filtering:

- **$namespace**: Filter by Kubernetes namespace
- **$policy**: Filter by EnforcementPolicy name

These allow drilling down into specific namespaces or policies.

## Business Metrics Guide

### For Executives

**Key question:** "What's the ROI?"

**Show this panel:** ROI Calculator (Panel 2)
- Example: 22.5x means for every $1 spent on this tool, you save $22.50

**Also useful:**
- Monthly Savings Trajectory (Panel 1) - "We're saving $X/month"
- Savings Over Time (Panel 6) - "Trend is improving"

### For Finance

**Key question:** "Where are the savings coming from?"

**Show this panel:** Top Cost-Saving Policies (Panel 7)
- Ranked list of which policies generate most savings
- Helps prioritize which environments to focus on

**Also useful:**
- Savings Over Time (Panel 6) - Monthly cost reduction
- Actions Per Day (Panel 5) - Volume of optimization

### For Engineering

**Key question:** "Is this causing problems?"

**Show this panel:** Enforcement Accuracy (Panel 3)
- >95% accuracy = minimal false positives
- <90% accuracy = policies need tuning

**Also useful:**
- Reactivation Rate (Panel 10) - How often people reverse actions
- Policy Violations & Overrides (Panel 9) - Manual exclusions
- Average Reconciliation Time (Panel 11) - Performance

### For DevOps

**Key question:** "Is the controller healthy?"

**Show this panel:** Average Reconciliation Time (Panel 11)
- <10s = healthy
- >30s = needs investigation (too many resources? OpenCost slow?)

**Also useful:**
- Resource Matches vs Actions (Panel 8) - Rate limiting impact
- Resources Currently Paused (Panel 4) - Current state

## Customization

### Change ROI Calculation

Edit Panel 2 query:
```promql
# Current: Assumes $200/month overhead
(sum(finops_estimated_savings_usd) * 12) / 2400

# Custom: Your actual overhead (e.g., $500/month)
(sum(finops_estimated_savings_usd) * 12) / 6000
```

### Add Cloud Cost Comparison

Add new panel comparing FinOps Enforcer savings to total cloud spend:

```promql
# Percentage of total cloud spend saved
(sum(finops_estimated_savings_usd) / YOUR_MONTHLY_CLOUD_SPEND) * 100
```

### Namespace-Specific Dashboard

Duplicate dashboard and hard-code namespace filter:
```promql
sum(finops_paused_resources_total{namespace="dev-team1"})
```

## Alerting

Recommended Prometheus alerts based on dashboard metrics:

```yaml
groups:
  - name: finops_enforcer
    rules:
      # Low accuracy
      - alert: FinOpsLowAccuracy
        expr: (1 - (sum(rate(finops_false_positives_total[24h])) / sum(rate(finops_actions_taken_total[24h])))) * 100 < 90
        for: 1h
        annotations:
          summary: "FinOps Enforcer accuracy below 90%"
          description: "Too many false positives - policies may need tuning"
      
      # Slow reconciliation
      - alert: FinOpsSlowReconciliation
        expr: avg(finops_reconciliation_duration_seconds) > 30
        for: 5m
        annotations:
          summary: "FinOps Enforcer reconciliation slow"
          description: "Reconciliation taking >30s - may need optimization"
      
      # High reactivation rate
      - alert: FinOpsHighReactivationRate
        expr: sum(increase(finops_reactivations_total[24h])) > 10
        for: 1h
        annotations:
          summary: "Many resources being reactivated"
          description: "Possible false positives or policies too aggressive"
```

## Troubleshooting

### Dashboard shows "No data"

1. **Check Prometheus data source:**
   ```bash
   kubectl port-forward -n monitoring svc/prometheus 9090:9090
   # Visit http://localhost:9090
   # Query: finops_paused_resources_total
   ```

2. **Verify controller metrics:**
   ```bash
   kubectl port-forward -n finops-system svc/finops-enforcer-metrics 8080:8080
   curl http://localhost:8080/metrics | grep finops_
   ```

3. **Check ServiceMonitor:**
   ```bash
   kubectl get servicemonitor -n finops-system
   kubectl describe servicemonitor finops-enforcer -n finops-system
   ```

### ROI panel shows unrealistic values

- Check overhead assumption in query (default: $2400/year)
- Adjust for your actual controller costs
- Consider including operator time (e.g., 2 hours/month @ $100/hr = $200/month)

### Savings seem low

- Many resources in dry-run mode?
- Check cost thresholds (might be too high)
- Verify OpenCost is returning accurate data
- Check namespace scope (policies covering right namespaces?)

## Screenshots

_(In production, add screenshots here of each panel)_

## Feedback

Have ideas for additional panels? Open an issue or PR!

Useful additions:
- Cost breakdown by namespace
- Comparison to cloud billing data
- Forecasted annual savings
- Policy effectiveness scores
