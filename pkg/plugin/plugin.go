package plugin

import "k8s.io/apimachinery/pkg/runtime"

var WorkloadInjectors []WorkloadInjector

func Register(wi WorkloadInjector) {
	WorkloadInjectors = append(WorkloadInjectors, wi)
}

type WorkloadInjector interface {
	Name() string

	Match(runtime.Object) bool

	Inject(runtime.Object) runtime.Object
}
