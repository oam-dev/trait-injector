package injector

import (
	"github.com/oam-dev/trait-injector/pkg/plugin"

	corev1 "k8s.io/api/core/v1"
)

func makeVolumeMountName(secretName, pvcName string) string {
	var s string
	if len(secretName) != 0 {
		s = "secret-" + secretName
	} else if len(pvcName) != 0 {
		s = "pvc-" + pvcName
	}
	return s
}

func getValues(ctx plugin.TargetContext) (string, string) {
	var secretName, pvcName string
	if val, ok := ctx.Values["secret-name"]; ok {
		secretName = val.(string)
	}
	if val, ok := ctx.Values["pvc-name"]; ok {
		pvcName = val.(string)
	}
	return secretName, pvcName
}

func makeVolumeSource(secretName, pvcName string) corev1.VolumeSource {
	var vs corev1.VolumeSource
	if len(secretName) != 0 {
		vs = corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		}
	} else if len(pvcName) != 0 {
		vs = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		}
	}
	return vs
}

func FindString(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
