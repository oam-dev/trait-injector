package injector

import (
	"github.com/oam-dev/trait-injector/pkg/plugin"
)

func Defaults() []plugin.TargetInjector {
	return []plugin.TargetInjector{
		newDeploymentTargetInjector(),
		newStatefulsetTargetInjector(),
	}
}
