package injector

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/go-logr/logr"
	"github.com/oam-dev/trait-injector/pkg/plugin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	volumeMountName = "service-binding-volume"
)

var _ plugin.TargetInjector = &DeploymentTargetInjector{}

type DeploymentTargetInjector struct {
	Log logr.Logger
}

func (ti *DeploymentTargetInjector) Name() string {
	return "DeploymentTargetInjector"
}

func (ti *DeploymentTargetInjector) Match(k metav1.GroupVersionKind) bool {
	if k.Group == "apps" && k.Version == "v1" && k.Kind == "Deployment" {
		return true
	}
	return false
}

func (ti *DeploymentTargetInjector) Inject(ctx plugin.TargetContext, raw runtime.RawExtension) ([]webhook.JSONPatchOp, error) {
	var deployment *appsv1.Deployment
	err := json.Unmarshal(raw.Raw, &deployment)
	if err != nil {
		return nil, err
	}

	var patches []webhook.JSONPatchOp

	b := ctx.Binding
	secretName, pvcName := getValues(ctx)
	// Inject secret to env in deployment
	if b.To.Env {
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if len(c.EnvFrom) == 0 {
				patch := webhook.JSONPatchOp{
					Operation: "add",
					Path:      fmt.Sprintf("/spec/template/spec/containers/%d/envFrom", i),
					Value:     []corev1.EnvFromSource{},
				}
				patches = append(patches, patch)
			}

			patch := webhook.JSONPatchOp{
				Operation: "add",
				Path:      fmt.Sprintf("/spec/template/spec/containers/%d/envFrom/-", i),
				Value: corev1.EnvFromSource{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
					},
				},
			}
			patches = append(patches, patch)
		}
		ti.Log.Info("injected secret to env", "deployment", path.Join(deployment.Namespace, deployment.Name))
	}

	// inject secret as file in Pod
	if len(b.To.FilePath) != 0 {
		if len(deployment.Spec.Template.Spec.Volumes) == 0 {
			patch := webhook.JSONPatchOp{
				Operation: "add",
				Path:      "/spec/template/spec/volumes",
				Value:     []corev1.Volume{},
			}
			patches = append(patches, patch)
		}

		patch := webhook.JSONPatchOp{
			Operation: "add",
			Path:      "/spec/template/spec/volumes/-",
			Value: corev1.Volume{
				Name:         volumeMountName,
				VolumeSource: makeVolumeSource(secretName, pvcName),
			},
		}
		patches = append(patches, patch)

		for i, c := range deployment.Spec.Template.Spec.Containers {
			if len(c.VolumeMounts) == 0 {
				patch := webhook.JSONPatchOp{
					Operation: "add",
					Path:      fmt.Sprintf("/spec/template/spec/containers/%d/volumeMounts", i),
					Value:     []corev1.VolumeMount{},
				}
				patches = append(patches, patch)
			}

			patch := webhook.JSONPatchOp{
				Operation: "add",
				Path:      fmt.Sprintf("/spec/template/spec/containers/%d/volumeMounts/-", i),
				Value: corev1.VolumeMount{
					Name:      volumeMountName,
					MountPath: b.To.FilePath,
				},
			}
			patches = append(patches, patch)
		}

		ti.Log.Info("injected volume to file", "deployment", path.Join(deployment.Namespace, deployment.Name))
	}

	return patches, nil
}
