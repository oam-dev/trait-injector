package injector

import (
	"testing"

	corev1alpha1 "github.com/oam-dev/trait-injector/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
