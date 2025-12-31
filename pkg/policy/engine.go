package policy
package policy

import (
	"context"
	"path/filepath"
	"time"

	finopsv1alpha1 "github.com/yourusername/finops-enforcer/api/v1alpha1"
	"github.com/yourusername/finops-enforcer/pkg/cost"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Engine evaluates enforcement policies against resources
type Engine struct {
	// Dependencies injected at initialization
}

// NewEngine creates a new policy engine
func NewEngine() *Engine {
	return &Engine{}
}

// EvaluationResult represents the result of policy evaluation
type EvaluationResult struct {
	Policy     *finopsv1alpha1.EnforcementPolicy
	Deployment *appsv1.Deployment
	CostData   *cost.CostData
	Matched    bool










































































































































































































































}	return metav1.FormatResourceQuantity(metav1.ResourceList{}, 2)func formatFloat(f float64) string {// formatFloat formats float with 2 decimal places}		", zero traffic detected, hourly cost: $" + formatFloat(costData.HourlyCost)	return "Idle for " + policy.Spec.Conditions.IdleWindow.Duration.String() +func buildMatchReason(policy *finopsv1alpha1.EnforcementPolicy, costData *cost.CostData) string {// buildMatchReason constructs a human-readable reason for policy match}	return matched	matched, _ := filepath.Match(pattern, value)func matchPattern(pattern, value string) bool {// matchPattern performs wildcard pattern matching}	return deployment.Annotations["finops.io/exclude"] == "true"func isExcluded(deployment *appsv1.Deployment) bool {// isExcluded checks if deployment has exclusion annotation}	return deployment.Annotations["finops.io/paused"] == "true"func isPaused(deployment *appsv1.Deployment) bool {// isPaused checks if deployment is already paused}	return time.Since(pausedAt) >= cooldownWindow	}		return true // Can't parse - assume expired	if err != nil {	pausedAt, err := time.Parse(time.RFC3339, pausedAtStr)	}		return true // Never paused before	if pausedAtStr == "" {	pausedAtStr := deployment.Annotations["finops.io/paused-at"]func (e *Engine) isCooldownExpired(deployment *appsv1.Deployment, cooldownWindow time.Duration) bool {// isCooldownExpired checks if cooldown period has passed}	return false	}		}			}				return true			if currentHour >= start && currentHour <= end {			start, end := activeHours.Hours[0], activeHours.Hours[1]		if len(activeHours.Hours) == 2 {		// Check if current hour is within range		}			continue		if !dayMatches {		}			}				break				dayMatches = true			if day == currentDay {		for _, day := range activeHours.Days {		dayMatches := false		// Check if current day matches	for _, activeHours := range schedule.ActiveHours {	currentHour := now.Hour()	currentDay := now.Weekday().String()[:3] // Mon, Tue, etc.	now := time.Now().In(loc)	}		return true		// Invalid timezone - default to allowing	if err != nil {	loc, err := time.LoadLocation(schedule.Timezone)func (e *Engine) isWithinSchedule(schedule *finopsv1alpha1.ScheduleSpec) bool {// isWithinSchedule checks if current time is within policy schedule}	return time.Since(lastActivity) >= idleWindow	}		return true	if err != nil {	lastActivity, err := time.Parse(time.RFC3339, lastActivityStr)	}		return true		// In production, this would integrate with metrics/traffic data		// No activity tracking - assume idle	if lastActivityStr == "" {	lastActivityStr := deployment.Annotations["finops.io/last-activity"]	// Check for last activity annotationfunc (e *Engine) isIdleLongEnough(deployment *appsv1.Deployment, idleWindow time.Duration) bool {// isIdleLongEnough checks if deployment has been idle for required duration}	return true	}		}			return false		if labels[key] != value {	for key, value := range filter.Match {	// Check required matches	}		}			return false		if labels[key] == value {	for key, value := range filter.Exclude {	// Check exclusions firstfunc (e *Engine) matchesLabelFilter(labels map[string]string, filter finopsv1alpha1.LabelFilter) bool {// matchesLabelFilter checks if labels match policy filter}	return false	}		}			return true		if matchPattern(pattern, namespace) {	for _, pattern := range filter.Include {	// Check inclusions	}		}			return false		if matchPattern(pattern, namespace) {	for _, pattern := range filter.Exclude {	// Check exclusions firstfunc (e *Engine) matchesNamespaceScope(namespace string, filter finopsv1alpha1.NamespaceFilter) bool {// matchesNamespaceScope checks if namespace matches policy scope}	return result, nil	}		DryRun:                  policy.Spec.Enforcement.DryRun,		Policy:                  policy.Name,		EstimatedMonthlySavings: cost.EstimateMonthlyCost(costData.HourlyCost),		Reason:                  result.Reason,		OriginalReplicas:        *deployment.Spec.Replicas,		Deployment:              deployment,		Type:                    policy.Spec.Actions.Type,	result.Action = &EnforcementAction{	result.Reason = buildMatchReason(policy, costData)	result.Matched = true	// All conditions matched - create action	}		}			return result, nil			result.Reason = "cooldown not expired"		if !e.isCooldownExpired(deployment, policy.Spec.Enforcement.CooldownWindow.Duration) {	if policy.Spec.Enforcement.CooldownWindow.Duration > 0 {	// Check cooldown	}		}			return result, nil			result.Reason = "outside scheduled hours"		if !e.isWithinSchedule(policy.Spec.Schedule) {	if policy.Spec.Schedule != nil {	// Check schedule (if defined)	}		return result, nil		result.Reason = "not idle long enough"	if !e.isIdleLongEnough(deployment, policy.Spec.Conditions.IdleWindow.Duration) {	// Check idle window	}		return result, nil		result.Reason = "cost below threshold"	if costData.HourlyCost < policy.Spec.Conditions.MinHourlyCost {	// Check cost threshold	}		}			return result, nil			result.Reason = "labels do not match"		if !e.matchesLabelFilter(deployment.Labels, *policy.Spec.Scope.Labels) {	if policy.Spec.Scope.Labels != nil {	// Check label filters	}		return result, nil		result.Reason = "namespace not in scope"	if !e.matchesNamespaceScope(deployment.Namespace, policy.Spec.Scope.Namespaces) {	// Check namespace scope	}		return result, nil		result.Reason = "excluded by annotation"	if isExcluded(deployment) {	// Check for exclusion annotation	}		return result, nil		result.Reason = "already paused"	if isPaused(deployment) {	// Check if already paused	}		Matched:    false,		CostData:   costData,		Deployment: deployment,		Policy:     policy,	result := &EvaluationResult{) (*EvaluationResult, error) {	costData *cost.CostData,	deployment *appsv1.Deployment,	policy *finopsv1alpha1.EnforcementPolicy,	ctx context.Context,func (e *Engine) Evaluate(// Evaluate evaluates a deployment against a policy}	DryRun                   bool	Policy                   string	EstimatedMonthlySavings  float64	Reason                   string	OriginalReplicas         int32	Deployment               *appsv1.Deployment	Type                     finopsv1alpha1.ActionTypetype EnforcementAction struct {// EnforcementAction represents an action to be taken}	Action     *EnforcementAction	Reason     string