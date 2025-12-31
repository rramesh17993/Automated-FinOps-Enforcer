package metrics
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// PausedResourcesTotal tracks the total number of currently paused resources
	PausedResourcesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "finops_paused_resources_total",
			Help: "Number of resources currently paused by FinOps Enforcer",
		},
		[]string{"namespace", "policy"},
	)

	// EstimatedSavingsUSD tracks estimated monthly savings
	EstimatedSavingsUSD = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "finops_estimated_savings_usd",
			Help: "Estimated monthly savings in USD from paused resources",
		},
		[]string{"namespace"},
	)

























































































































}	FalsePositivesTotal.WithLabelValues(namespace, policy).Inc()func RecordFalsePositive(namespace, policy string) {// RecordFalsePositive increments false positive counter}	ActionsTakenTotal.WithLabelValues(action, namespace, dryRunStr).Inc()	}		dryRunStr = "true"	if dryRun {	dryRunStr := "false"func RecordAction(action, namespace string, dryRun bool) {// RecordAction increments action counter}	PolicyMatchesTotal.WithLabelValues(policy, action).Inc()func RecordPolicyMatch(policy, action string) {// RecordPolicyMatch increments policy match counter}	EstimatedSavingsUSD.WithLabelValues(namespace).Sub(savings)	PausedResourcesTotal.WithLabelValues(namespace, "").Dec()	ReactivationsTotal.WithLabelValues(namespace, source).Inc()func RecordReactivation(namespace, source string, savings float64) {// RecordReactivation increments reactivation metric}	EstimatedSavingsUSD.WithLabelValues(namespace).Add(savings)	PausedResourcesTotal.WithLabelValues(namespace, policy).Inc()func RecordPausedResource(namespace, policy string, savings float64) {// RecordPausedResource increments paused resources metric}	)		PolicyEvaluationErrors,		OpenCostAPIErrors,		ReconciliationDuration,		PolicyEvaluationDuration,		FalsePositivesTotal,		ReactivationsTotal,		ActionsTakenTotal,		PolicyMatchesTotal,		EstimatedSavingsUSD,		PausedResourcesTotal,	metrics.Registry.MustRegister(	// Register metrics with controller-runtimefunc init() {)	)		[]string{"policy"},		},			Help: "Number of policy evaluation errors",			Name: "finops_policy_evaluation_errors_total",		prometheus.CounterOpts{	PolicyEvaluationErrors = prometheus.NewCounterVec(	// PolicyEvaluationErrors tracks policy evaluation failures	)		},			Help: "Number of OpenCost API errors encountered",			Name: "finops_opencost_api_errors_total",		prometheus.CounterOpts{	OpenCostAPIErrors = prometheus.NewCounter(	// OpenCostAPIErrors tracks OpenCost API failures	)		},			Buckets: prometheus.DefBuckets,			Help:    "Time spent in reconciliation loop",			Name:    "finops_reconciliation_duration_seconds",		prometheus.HistogramOpts{	ReconciliationDuration = prometheus.NewHistogram(	// ReconciliationDuration tracks overall reconciliation loop duration	)		[]string{"policy"},		},			Buckets: prometheus.DefBuckets,			Help:    "Time spent evaluating policies",			Name:    "finops_policy_evaluation_duration_seconds",		prometheus.HistogramOpts{	PolicyEvaluationDuration = prometheus.NewHistogramVec(	// PolicyEvaluationDuration tracks time spent evaluating policies	)		[]string{"namespace", "policy"},		},			Help: "Resources reactivated within 1 hour (likely false positive)",			Name: "finops_false_positives_total",		prometheus.CounterOpts{	FalsePositivesTotal = prometheus.NewCounterVec(	// FalsePositivesTotal tracks resources reactivated within 1 hour	)		[]string{"namespace", "source"},		},			Help: "Number of user-initiated reactivations",			Name: "finops_reactivations_total",		prometheus.CounterOpts{	ReactivationsTotal = prometheus.NewCounterVec(	// ReactivationsTotal counts user-initiated reactivations	)		[]string{"action", "namespace", "dry_run"},		},			Help: "Number of enforcement actions taken",			Name: "finops_actions_taken_total",		prometheus.CounterOpts{	ActionsTakenTotal = prometheus.NewCounterVec(	// ActionsTakenTotal counts enforcement actions executed	)		[]string{"policy", "action"},		},			Help: "Number of times policies matched resources",			Name: "finops_policy_matches_total",		prometheus.CounterOpts{	PolicyMatchesTotal = prometheus.NewCounterVec(	// PolicyMatchesTotal counts policy evaluation matches