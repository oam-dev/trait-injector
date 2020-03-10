package plugin

import (
	corev1alpha1 "github.com/oam-dev/trait-injector/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var TargetInjectors []TargetInjector

func RegisterTargetInjectors(ts ...TargetInjector) {
	TargetInjectors = append(TargetInjectors, ts...)
}

// TargetInjector handles data injection to workload target.
type TargetInjector interface {
	Name() string

	Match(metav1.GroupVersionKind) bool

	Inject(TargetContext, runtime.RawExtension) ([]webhook.JSONPatchOp, error)
}

type TargetContext struct {
	Binding *corev1alpha1.Binding
	Values  map[string]interface{}
}
