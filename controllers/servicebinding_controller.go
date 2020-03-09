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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/oam-dev/trait-injector/api/v1alpha1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.oam.dev,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.oam.dev,resources=servicebindings/status,verbs=get;update;patch

func (r *ServiceBindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("servicebinding", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ServiceBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.ServiceBinding{}).
		Complete(r)
}

func (r *ServiceBindingReconciler) ServeAdmission() {
	mux := http.NewServeMux()

	mux.HandleFunc("/mutate", r.handleMutate)

	port := ":8443"
	s := &http.Server{
		Addr:           port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}
	r.Log.Info("listening on", "port", port)
	panic(s.ListenAndServeTLS("./ssl/service-injector.pem", "./ssl/service-injector.key"))
}

func (r *ServiceBindingReconciler) handleMutate(w http.ResponseWriter, req *http.Request) {
	err := r.handleMutateErr(w, req)
	if err != nil {
		r.Log.Error(err, "HandleMutate")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}
}

func (r *ServiceBindingReconciler) handleMutateErr(w http.ResponseWriter, req *http.Request) error {
	// parse AdmissionReview
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return fmt.Errorf("read request body err: %w", err)
	}

	review := &admissionv1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, review); err != nil {
		return fmt.Errorf("unmarshal AdmissionReview err: %w", err)
	}

	patches, err := r.handleAdmissionRequest(review.Request)
	if err != nil {
		return fmt.Errorf("handleAdmissionRequest err: %w", err)
	}

	// write back response
	p, err := json.Marshal(patches)
	if err != nil {
		return err
	}
	review.Response = newAdmissionResponse(review, p)
	b, err := json.Marshal(review)
	if err != nil {
		return fmt.Errorf("marshal AdmissionReview err: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}

func newAdmissionResponse(review *admissionv1beta1.AdmissionReview, patch []byte) *admissionv1beta1.AdmissionResponse {
	resp := &admissionv1beta1.AdmissionResponse{}
	// set response options
	resp.Allowed = true
	resp.UID = review.Request.UID
	pT := admissionv1beta1.PatchTypeJSONPatch
	resp.PatchType = &pT

	resp.Patch = patch
	resp.Result = &metav1.Status{
		Status: "Success",
	}

	return resp
}

func (r *ServiceBindingReconciler) handleAdmissionRequest(req *admissionv1beta1.AdmissionRequest) ([]webhook.JSONPatchOp, error) {
	// Search any ServiceBinding whose target matches the given request.
	sbl := &corev1alpha1.ServiceBindingList{}
	err := r.Client.List(context.TODO(), sbl)
	if err != nil {
		return nil, fmt.Errorf("list servicebindings err: %w", err)
	}

	var sb *corev1alpha1.ServiceBinding
	for _, item := range sbl.Items {
		obj := item.Spec.WorkloadRef
		r.Log.Info("kind matching", "apiVersion", obj.APIVersion, "kind", obj.Kind, "request", req.Kind.String())
		rk := req.Kind
		gv := rk.Version
		if len(rk.Group) > 0 {
			gv = fmt.Sprintf("%s/%s", rk.Group, rk.Version)
		}
		if rk.Kind == obj.Kind && gv == obj.APIVersion {
			sb = &item
			break
		}
	}
	if sb == nil {
		r.Log.Info("uninterested request", "request", path.Join(req.Namespace, req.Name))
		return nil, nil
	}

	for _, b := range sb.Spec.Bindings {
		if b.From.Secret != nil {
			return r.injectSecret(req, sb, b)
		}
	}
	return nil, nil
}

func (r *ServiceBindingReconciler) injectSecret(req *admissionv1beta1.AdmissionRequest, sb *corev1alpha1.ServiceBinding, b corev1alpha1.Binding) ([]webhook.JSONPatchOp, error) {
	w := sb.Spec.WorkloadRef
	s := b.From.Secret

	secretName := s.Name
	if s.NameFromField != nil {
		// TODO: use dynamic client to read
	}

	switch {
	// Deployment target
	case w.APIVersion == "apps/v1" && w.Kind == "Deployment":
		var deployment *appsv1.Deployment
		err := json.Unmarshal(req.Object.Raw, &deployment)
		if err != nil {
			return nil, err
		}

		var patches []webhook.JSONPatchOp

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
			r.Log.Info("injected secret to env", "deployment", path.Join(deployment.Namespace, deployment.Name))
		}

		return patches, nil
	case w.APIVersion == "apps/v1" && w.Kind == "StatefulSet":
		panic("TODO: support StatefulSet")
	default:
		r.Log.Info("unsupported target kind ", "apiVersion", w.APIVersion, "kind", w.Kind)
		return nil, nil
	}
}
