package plugin

import (
	corev1alpha1 "github.com/oam-dev/trait-injector/api/v1alpha1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
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

	Match(*admissionv1beta1.AdmissionRequest, *corev1alpha1.WorkloadReference) bool

	Inject(TargetContext, runtime.RawExtension) ([]webhook.JSONPatchOp, error)
}

type TargetContext struct {
	Binding *corev1alpha1.Binding
	Values  map[string]interface{}
}
