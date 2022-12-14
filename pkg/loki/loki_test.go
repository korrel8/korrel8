package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/korrel8/korrel8/internal/pkg/test"
	"github.com/korrel8/korrel8/pkg/korrel8"
	"github.com/korrel8/korrel8/pkg/uri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ctx = context.Background()

func TestStore_Get_PlainLoki(t *testing.T) {
	t.Parallel()
	l := test.RequireLokiServer(t)
	lines := []string{"hello", "there", "mr. frog"}
	err := l.Push(map[string]string{"test": "loki"}, lines...)
	require.NoError(t, err)
	s, err := NewStore(l.URL(), http.DefaultClient)
	require.NoError(t, err)

	var want []korrel8.Object
	for _, l := range lines {
		want = append(want, Object(l))
	}
	ref := NewPlainRef(`{test="loki"}`, nil)
	result := korrel8.NewListResult()
	require.NoError(t, s.Get(ctx, ref, result))
	assert.Equal(t, want, result.List())
}

func TestLokiStackStore_Get(t *testing.T) {
	t.Parallel()
	test.SkipIfNoCluster(t)
	c := test.K8sClient
	ns := test.TempNamespace(t, c)
	want := []string{"hello", "there", "mr.", "frog"}

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "logger", Namespace: ns},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    "logger",
				Image:   "quay.io/quay/busybox",
				Command: []string{"sh", "-c", fmt.Sprintf("echo %v; sleep infinity", strings.Join(want, "; echo "))}}}},
	}
	require.NoError(t, c.Create(ctx, &pod))
	s, err := NewOpenshiftLokiStackStore(ctx, c, test.RESTConfig)
	require.NoError(t, err)
	logQL := fmt.Sprintf(`{kubernetes_pod_name="%v", kubernetes_namespace_name="%v"}`, pod.Name, pod.Namespace)
	ref := NewLokiStackRef(Application, logQL, nil)
	var result korrel8.ListResult
	assert.Eventually(t, func() bool {
		result = nil
		err = s.Get(ctx, ref, &result)
		require.NoError(t, err)
		t.Logf("waiting for 4 logs, got %v. %v%v", len(result), s, ref)
		return len(result) >= 3
	}, time.Minute, 5*time.Second)
	var got []string
	for _, obj := range result {
		var m map[string]any
		line := string(obj.(Object))
		assert.NoError(t, json.Unmarshal([]byte(line), &m), line)
		got = append(got, m["message"].(string))
	}
	assert.Equal(t, want, got)
}

func TestStoreGet_Constraint(t *testing.T) {
	t.Parallel()
	l := test.RequireLokiServer(t)

	err := l.Push(map[string]string{"test": "loki"}, "much", "too", "early")
	require.NoError(t, err)

	t1 := time.Now()
	err = l.Push(map[string]string{"test": "loki"}, "right", "on", "time")
	require.NoError(t, err)
	t2 := time.Now()

	err = l.Push(map[string]string{"test": "loki"}, "much", "too", "late")
	require.NoError(t, err)
	s, err := NewStore(l.URL(), http.DefaultClient)
	require.NoError(t, err)

	for n, x := range []struct {
		ref  uri.Reference
		want []korrel8.Object
	}{
		{
			ref:  NewPlainRef(`{test="loki"}`, &korrel8.Constraint{End: &t1}),
			want: []korrel8.Object{Object("much"), Object("too"), Object("early")},
		},
		{
			ref:  NewPlainRef(`{test="loki"}`, &korrel8.Constraint{Start: &t1, End: &t2}),
			want: []korrel8.Object{Object("right"), Object("on"), Object("time")},
		},
	} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			var result korrel8.ListResult
			assert.NoError(t, s.Get(ctx, x.ref, &result))
			assert.Equal(t, x.want, result.List())
		})
	}
}
