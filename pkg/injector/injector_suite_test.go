package injector

import (
	"encoding/json"
	"testing"

	corev1alpha1 "github.com/oam-dev/trait-injector/api/v1alpha1"
	"github.com/oam-dev/trait-injector/pkg/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func TestInjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Injector Suite")
}

var _ = Describe("Injector", func() {
	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	di := newDeploymentTargetInjector()
	si := newStatefulsetTargetInjector()

	Describe("workload injection", func() {
		It("should inject secret to Deployment env", func() {
			ctx := plugin.TargetContext{
				Binding: &corev1alpha1.Binding{
					From: corev1alpha1.DataSource{
						Secret: &corev1alpha1.SecretSource{
							Name: "test-secret",
						},
					},
					To: corev1alpha1.DataTarget{
						Env: true,
					},
				},
				Values: map[string]interface{}{"secret-name": "test-secret"},
			}
			d := &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deploy",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "test-container",
							}},
						},
					},
				},
			}
			b, err := json.Marshal(d)
			Expect(err).To(BeNil())
			raw := runtime.RawExtension{
				Raw: b,
			}

			patches, err := di.Inject(ctx, raw)
			Expect(err).To(BeNil())
			Expect(patches).To(Equal([]webhook.JSONPatchOp{{
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/envFrom",
				Value:     []corev1.EnvFromSource{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/envFrom/-",
				Value: corev1.EnvFromSource{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
					},
				},
			}}))
		})
		It("should inject secret to StatefulSet env", func() {
			ctx := plugin.TargetContext{
				Binding: &corev1alpha1.Binding{
					From: corev1alpha1.DataSource{
						Secret: &corev1alpha1.SecretSource{
							Name: "test-secret",
						},
					},
					To: corev1alpha1.DataTarget{
						Env: true,
					},
				},
				Values: map[string]interface{}{"secret-name": "test-secret"},
			}
			d := &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deploy",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "test-container",
							}},
						},
					},
				},
			}
			b, err := json.Marshal(d)
			Expect(err).To(BeNil())
			raw := runtime.RawExtension{
				Raw: b,
			}

			patches, err := si.Inject(ctx, raw)
			Expect(err).To(BeNil())
			Expect(patches).To(Equal([]webhook.JSONPatchOp{{
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/envFrom",
				Value:     []corev1.EnvFromSource{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/envFrom/-",
				Value: corev1.EnvFromSource{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
					},
				},
			}}))
		})
		It("should inject secret to Deployment filePath", func() {
			ctx := plugin.TargetContext{
				Binding: &corev1alpha1.Binding{
					From: corev1alpha1.DataSource{
						Secret: &corev1alpha1.SecretSource{
							Name: "test-secret",
						},
					},
					To: corev1alpha1.DataTarget{
						FilePath: "/test/path",
					},
				},
				Values: map[string]interface{}{"secret-name": "test-secret"},
			}
			d := &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deploy",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "test-container",
							}},
						},
					},
				},
			}
			b, err := json.Marshal(d)
			Expect(err).To(BeNil())
			raw := runtime.RawExtension{
				Raw: b,
			}

			patches, err := di.Inject(ctx, raw)
			Expect(err).To(BeNil())
			Expect(patches).To(Equal([]webhook.JSONPatchOp{{
				Operation: "add",
				Path:      "/spec/template/spec/volumes",
				Value:     []corev1.Volume{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/volumes/-",
				Value: corev1.Volume{
					Name: "secret-test-secret",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret",
						},
					},
				},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts",
				Value:     []corev1.VolumeMount{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts/-",
				Value: corev1.VolumeMount{
					Name:      "secret-test-secret",
					MountPath: "/test/path",
				},
			}}))
		})
		It("should inject secret to StatefulSet filePath", func() {
			ctx := plugin.TargetContext{
				Binding: &corev1alpha1.Binding{
					From: corev1alpha1.DataSource{
						Secret: &corev1alpha1.SecretSource{
							Name: "test-secret",
						},
					},
					To: corev1alpha1.DataTarget{
						FilePath: "/test/path",
					},
				},
				Values: map[string]interface{}{"secret-name": "test-secret"},
			}
			d := &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deploy",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "test-container",
							}},
						},
					},
				},
			}
			b, err := json.Marshal(d)
			Expect(err).To(BeNil())
			raw := runtime.RawExtension{
				Raw: b,
			}

			patches, err := di.Inject(ctx, raw)
			Expect(err).To(BeNil())
			Expect(patches).To(Equal([]webhook.JSONPatchOp{{
				Operation: "add",
				Path:      "/spec/template/spec/volumes",
				Value:     []corev1.Volume{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/volumes/-",
				Value: corev1.Volume{
					Name: "secret-test-secret",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret",
						},
					},
				},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts",
				Value:     []corev1.VolumeMount{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts/-",
				Value: corev1.VolumeMount{
					Name:      "secret-test-secret",
					MountPath: "/test/path",
				},
			}}))
		})
		It("should inject pvc to Deployment filePath", func() {
			ctx := plugin.TargetContext{
				Binding: &corev1alpha1.Binding{
					From: corev1alpha1.DataSource{
						Volume: &corev1alpha1.VolumeSource{
							PVCName: "test-PVC",
						},
					},
					To: corev1alpha1.DataTarget{
						FilePath: "/test/path",
					},
				},
				Values: map[string]interface{}{"pvc-name": "test-pvc"},
			}
			d := &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deploy",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "test-container",
							}},
						},
					},
				},
			}
			b, err := json.Marshal(d)
			Expect(err).To(BeNil())
			raw := runtime.RawExtension{
				Raw: b,
			}

			patches, err := di.Inject(ctx, raw)
			Expect(err).To(BeNil())
			Expect(patches).To(Equal([]webhook.JSONPatchOp{{
				Operation: "add",
				Path:      "/spec/template/spec/volumes",
				Value:     []corev1.Volume{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/volumes/-",
				Value: corev1.Volume{
					Name: "pvc-test-pvc",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "test-pvc",
						},
					},
				},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts",
				Value:     []corev1.VolumeMount{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts/-",
				Value: corev1.VolumeMount{
					Name:      "pvc-test-pvc",
					MountPath: "/test/path",
				},
			}}))
		})
		It("should inject pvc to StatefulSet filePath", func() {
			ctx := plugin.TargetContext{
				Binding: &corev1alpha1.Binding{
					From: corev1alpha1.DataSource{
						Volume: &corev1alpha1.VolumeSource{
							PVCName: "test-PVC",
						},
					},
					To: corev1alpha1.DataTarget{
						FilePath: "/test/path",
					},
				},
				Values: map[string]interface{}{"pvc-name": "test-pvc"},
			}
			d := &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deploy",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "test-container",
							}},
						},
					},
				},
			}
			b, err := json.Marshal(d)
			Expect(err).To(BeNil())
			raw := runtime.RawExtension{
				Raw: b,
			}

			patches, err := di.Inject(ctx, raw)
			Expect(err).To(BeNil())
			Expect(patches).To(Equal([]webhook.JSONPatchOp{{
				Operation: "add",
				Path:      "/spec/template/spec/volumes",
				Value:     []corev1.Volume{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/volumes/-",
				Value: corev1.Volume{
					Name: "pvc-test-pvc",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "test-pvc",
						},
					},
				},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts",
				Value:     []corev1.VolumeMount{},
			}, {
				Operation: "add",
				Path:      "/spec/template/spec/containers/0/volumeMounts/-",
				Value: corev1.VolumeMount{
					Name:      "pvc-test-pvc",
					MountPath: "/test/path",
				},
			}}))
		})
	})

	Describe("request matching", func() {
		It("should match Deployment injector", func() {
			req := &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				Name: "example",
			}
			wref := &corev1alpha1.WorkloadReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "example",
			}

			Expect(di.Match(req, wref)).To(Equal(true))
		})

		It("should not match Deployment injector if name mismatch", func() {
			req := &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				Name: "unmatch-name",
			}
			wref := &corev1alpha1.WorkloadReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "example",
			}

			Expect(di.Match(req, wref)).To(Equal(false))
		})

		It("should not match Deployment injector if gvk mismatch", func() {
			req := &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "UnmatchKind",
				},
				Name: "example",
			}
			wref := &corev1alpha1.WorkloadReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "example",
			}

			Expect(di.Match(req, wref)).To(Equal(false))
		})

		It("should match StatefulSet injector", func() {
			req := &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "StatefulSet",
				},
				Name: "example",
			}
			wref := &corev1alpha1.WorkloadReference{
				APIVersion: "apps/v1",
				Kind:       "StatefulSet",
				Name:       "example",
			}

			Expect(si.Match(req, wref)).To(Equal(true))
		})

		It("should not match StatefulSet injector if name mismatch", func() {
			req := &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "StatefulSet",
				},
				Name: "unmatch-name",
			}
			wref := &corev1alpha1.WorkloadReference{
				APIVersion: "apps/v1",
				Kind:       "StatefulSet",
				Name:       "example",
			}

			Expect(si.Match(req, wref)).To(Equal(false))
		})

		It("should not match StatefulSet injector if gvk mismatch", func() {
			req := &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "UnmatchKind",
				},
				Name: "example",
			}
			wref := &corev1alpha1.WorkloadReference{
				APIVersion: "apps/v1",
				Kind:       "StatefulSet",
				Name:       "example",
			}

			Expect(si.Match(req, wref)).To(Equal(false))
		})

	})
})
