package policy

import (
	"testing"

	finopsv1alpha1 "github.com/yourusername/finops-enforcer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMatchesNamespaceScope(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name      string
		namespace string
		filter    finopsv1alpha1.NamespaceFilter
		want      bool
	}{
		{
			name:      "exact match",
			namespace: "dev-test",
			filter: finopsv1alpha1.NamespaceFilter{
				Include: []string{"dev-test"},
			},
			want: true,
		},
		{
			name:      "wildcard match",
			namespace: "dev-feature-xyz",
			filter: finopsv1alpha1.NamespaceFilter{
				Include: []string{"dev-*"},
			},
			want: true,
		},
		{
			name:      "excluded namespace",
			namespace: "prod",
			filter: finopsv1alpha1.NamespaceFilter{
				Include: []string{"*"},
				Exclude: []string{"prod"},
			},
			want: false,
		},
		{
			name:      "no match",
			namespace: "staging",
			filter: finopsv1alpha1.NamespaceFilter{
				Include: []string{"dev-*"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.matchesNamespaceScope(tt.namespace, tt.filter)
			if got != tt.want {
				t.Errorf("matchesNamespaceScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPaused(t *testing.T) {
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		want       bool
	}{
		{
			name: "paused deployment",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"finops.io/paused": "true",
					},
				},
			},
			want: true,
		},
		{
			name: "not paused",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			want: false,
		},
		{
			name: "nil annotations",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPaused(tt.deployment)
			if got != tt.want {
				t.Errorf("isPaused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		want       bool
	}{
		{
			name: "excluded deployment",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"finops.io/exclude": "true",
					},
				},
			},
			want: true,
		},
		{
			name: "not excluded",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExcluded(tt.deployment)
			if got != tt.want {
				t.Errorf("isExcluded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{pattern: "dev-*", value: "dev-test", want: true},
		{pattern: "dev-*", value: "prod-test", want: false},
		{pattern: "*", value: "anything", want: true},
		{pattern: "exact", value: "exact", want: true},
		{pattern: "exact", value: "not-exact", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.value)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}
