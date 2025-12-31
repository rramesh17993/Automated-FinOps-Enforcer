package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnforcementPolicySpec defines the desired state of EnforcementPolicy
type EnforcementPolicySpec struct {
	// Scope defines which resources this policy applies to
	Scope ScopeSpec `json:"scope"`

	// Conditions define what qualifies as "idle"
	Conditions ConditionsSpec `json:"conditions"`

	// Actions define what to do when conditions are met
	Actions ActionsSpec `json:"actions"`

	// Enforcement contains enforcement constraints
	// +optional
	Enforcement EnforcementSpec `json:"enforcement,omitempty"`

	// Schedule defines when this policy is active
	// +optional
	Schedule *ScheduleSpec `json:"schedule,omitempty"`
}

// ScopeSpec defines the scope of resources to evaluate
type ScopeSpec struct {
	// Namespaces defines namespace filters
	Namespaces NamespaceFilter `json:"namespaces"`

	// Labels defines label-based filters
	// +optional
	Labels *LabelFilter `json:"labels,omitempty"`
}

// NamespaceFilter defines namespace inclusion/exclusion
type NamespaceFilter struct {
	// Include is a list of namespace patterns to include (supports wildcards)
	Include []string `json:"include"`

	// Exclude is a list of namespace patterns to exclude (supports wildcards)
	// +optional
	Exclude []string `json:"exclude,omitempty"`
}

// LabelFilter defines label-based filtering
type LabelFilter struct {
	// Match defines labels that must be present
	// +optional
	Match map[string]string `json:"match,omitempty"`

	// Exclude defines labels that exclude resources
	// +optional
	Exclude map[string]string `json:"exclude,omitempty"`
}

// ConditionsSpec defines idle detection criteria
type ConditionsSpec struct {
	// IdleWindow is the duration a resource must be idle before action
	IdleWindow metav1.Duration `json:"idleWindow"`

	// MinHourlyCost is the minimum hourly cost threshold
	MinHourlyCost float64 `json:"minHourlyCost"`

	// TrafficThreshold defines traffic-based idle detection
	// +optional
	TrafficThreshold *TrafficThresholdSpec `json:"trafficThreshold,omitempty"`

	// UtilizationThreshold defines resource utilization thresholds
	// +optional
	UtilizationThreshold *UtilizationThresholdSpec `json:"utilizationThreshold,omitempty"`
}

// TrafficThresholdSpec defines traffic-based idle criteria
type TrafficThresholdSpec struct {
	// RequestsPerMinute is the maximum requests/min to consider idle
	RequestsPerMinute int `json:"requestsPerMinute"`
}

// UtilizationThresholdSpec defines resource utilization thresholds
type UtilizationThresholdSpec struct {
	// CPU is the maximum CPU utilization percentage to consider idle
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory is the maximum memory utilization percentage to consider idle
	// +optional
	Memory string `json:"memory,omitempty"`
}

// ActionsSpec defines enforcement actions
type ActionsSpec struct {
	// Type is the action type (only "scaleToZero" supported)
	Type ActionType `json:"type"`

	// Notify defines notification method
	Notify NotifyType `json:"notify"`

	// ReactivationAllowed enables user-initiated reactivation
	ReactivationAllowed bool `json:"reactivationAllowed"`
}

// ActionType defines the type of enforcement action
// +kubebuilder:validation:Enum=scaleToZero
type ActionType string

const (
	ActionTypeScaleToZero ActionType = "scaleToZero"
)

// NotifyType defines notification method
// +kubebuilder:validation:Enum=slack;none
type NotifyType string

const (
	NotifyTypeSlack NotifyType = "slack"
	NotifyTypeNone  NotifyType = "none"
)

// EnforcementSpec defines enforcement constraints
type EnforcementSpec struct {
	// DryRun enables dry-run mode (no actual enforcement)
	// +optional
	DryRun bool `json:"dryRun,omitempty"`

	// MaxActionsPerRun limits actions per reconciliation
	// +optional
	MaxActionsPerRun int `json:"maxActionsPerRun,omitempty"`

	// CooldownWindow is the minimum time between actions on same resource
	// +optional
	CooldownWindow metav1.Duration `json:"cooldownWindow,omitempty"`
}

// ScheduleSpec defines when a policy is active
type ScheduleSpec struct {
	// Timezone for schedule interpretation (e.g., "America/Los_Angeles")
	Timezone string `json:"timezone"`

	// ActiveHours defines when policy is active
	ActiveHours []ActiveHoursSpec `json:"activeHours"`
}

// ActiveHoursSpec defines active time windows
type ActiveHoursSpec struct {
	// Days of week when active (Mon, Tue, Wed, Thu, Fri, Sat, Sun)
	Days []string `json:"days"`

	// Hours range when active [start, end] in 24-hour format
	Hours []int `json:"hours"`
}

// EnforcementPolicyStatus defines the observed state of EnforcementPolicy
type EnforcementPolicyStatus struct {
	// LastEvaluationTime is when policy was last evaluated
	// +optional
	LastEvaluationTime *metav1.Time `json:"lastEvaluationTime,omitempty"`

	// MatchedResources is the count of resources matching this policy
	// +optional
	MatchedResources int `json:"matchedResources,omitempty"`

	// ActionsPerformed is the count of actions taken by this policy
	// +optional
	ActionsPerformed int `json:"actionsPerformed,omitempty"`

	// EstimatedSavings is the estimated monthly savings in USD
	// +optional
	EstimatedSavings float64 `json:"estimatedSavings,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Matched",type=integer,JSONPath=`.status.matchedResources`
// +kubebuilder:printcolumn:name="Actions",type=integer,JSONPath=`.status.actionsPerformed`
// +kubebuilder:printcolumn:name="Savings",type=string,JSONPath=`.status.estimatedSavings`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// EnforcementPolicy is the Schema for the enforcementpolicies API
type EnforcementPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnforcementPolicySpec   `json:"spec,omitempty"`
	Status EnforcementPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EnforcementPolicyList contains a list of EnforcementPolicy
type EnforcementPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnforcementPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnforcementPolicy{}, &EnforcementPolicyList{})
}
