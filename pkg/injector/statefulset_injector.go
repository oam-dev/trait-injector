package injector

import (
	"github.com/go-logr/logr"
	"github.com/oam-dev/trait-injector/pkg/plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ plugin.TargetInjector = &StatefulsetTargetInjector{}

type StatefulsetTargetInjector struct {
	Log logr.Logger
}

func (ti *StatefulsetTargetInjector) Name() string {
	return "StatefulsetTargetInjector"
}

func (ti *StatefulsetTargetInjector) Match(k metav1.GroupVersionKind) bool {
	if k.Group == "apps" && k.Version == "v1" && k.Kind == "StatefulSet" {
		return true
	}
	return false
}

func (ti *StatefulsetTargetInjector) Inject(ctx plugin.TargetContext, raw runtime.RawExtension) ([]webhook.JSONPatchOp, error) {
}
