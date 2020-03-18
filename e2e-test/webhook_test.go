/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	corev1alpha1 "github.com/oam-dev/trait-injector/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Context("with a secret", func() {
	setupClient()
	ctx := context.Background()
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		StringData: map[string]string{"foo": "bar"},
		Type:       corev1.SecretTypeOpaque,
	}
	k8sClient.Create(ctx, s)

	Describe("with a service binding from secret to env", func() {
		sb := &corev1alpha1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-env",
				Namespace: "default",
			},
			Spec: corev1alpha1.ServiceBindingSpec{
				Bindings: []corev1alpha1.Binding{{
					From: corev1alpha1.DataSource{
						Secret: &corev1alpha1.SecretSource{
							Name: "test-secret",
						},
					},
					To: corev1alpha1.DataTarget{
						Env: true,
					},
				}},
				WorkloadRef: &corev1alpha1.WorkloadReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deploy",
				},
			},
		}
		k8sClient.Create(ctx, sb)

		It("should inject env to deployment", func() {
			d := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy",
					Namespace: "default",
					Labels:    map[string]string{"project": "oam-service-binding"},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "test-pod",
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:            "test-container",
								Image:           "busybox",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Command:         []string{"/bin/sh", "-c", "printenv; sleep 6000"},
							}},
						},
					},
				},
			}
			k8sClient.Create(ctx, d)

			By("checking deployment has env")
			objectKey := client.ObjectKey{
				Name:      d.Name,
				Namespace: d.Namespace,
			}
			res := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, objectKey, res),
				time.Second*5, time.Millisecond*500).Should(BeNil())

			Expect(len(res.Spec.Template.Spec.Containers[0].EnvFrom)).NotTo(BeZero())
		})
	})
})

var k8sClient client.Client

func setupClient() {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	corev1alpha1.AddToScheme(scheme)
	var err error
	k8sClient, err = client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		logf.Log.Error(err, "failed to create k8sClient")
		Fail("setup failed")
	}
}

func getResourceFunc(ctx context.Context, key client.ObjectKey, obj runtime.Object) func() error {
	return func() error {
		return k8sClient.Get(ctx, key, obj)
	}
}
