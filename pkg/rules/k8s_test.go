package rules

import (
	"testing"

	"github.com/alanconway/korrel8/pkg/k8s"
	"github.com/alanconway/korrel8/pkg/korrel8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func findRule(start, goal korrel8.Class) korrel8.Rule {
	for _, r := range K8sToK8s() {
		if r.Start() == start && r.Goal() == goal {
			return r
		}
	}
	return nil
}

func TestRules_DeploymentHasPods(t *testing.T) {
	r := findRule(k8s.ClassOf(&appsv1.Deployment{}), k8s.ClassOf(&corev1.Pod{}))
	require.NotNil(t, r)
	for _, x := range []struct {
		deployment *appsv1.Deployment
		query      string
	}{
		{
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "x"},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}},
				}},
			query: "/api/v1/namespaces/x/pods?labelSelector=app%3Dfoo",
		},
		{
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "y"},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}},
				}},
			query: "/api/v1/namespaces/y/pods?labelSelector=app%3Dfoo",
		},
	} {
		t.Run(x.query, func(t *testing.T) {
			result, err := r.Apply(k8s.Object{x.deployment}, nil)
			require.NoError(t, err)
			assert.Len(t, result, 1)
			assert.Equal(t, x.query, result[0])
		})
	}
}