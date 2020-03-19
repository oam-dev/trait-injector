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
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/oam-dev/trait-injector/api/v1alpha1"
	"github.com/oam-dev/trait-injector/pkg/plugin"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	Client   client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/", healthCheck)
	go http.ListenAndServe(":8888", healthMux)
	mux := http.NewServeMux()

	mux.HandleFunc("/mutate", r.handleMutate)
	//mux.HandleFunc("/", healthCheck)

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
	err := r.Client.List(context.TODO(), sbl, client.InNamespace(req.Namespace))
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
		switch {
		case b.From.Secret != nil:
			return r.injectSecret(req, sb.Spec.WorkloadRef, b)
		case b.From.Volume != nil:
			return r.injectVolume(req, sb.Spec.WorkloadRef, b)
		}
	}
	return nil, nil
}

func (r *ServiceBindingReconciler) injectVolume(req *admissionv1beta1.AdmissionRequest, w *corev1alpha1.WorkloadReference, b corev1alpha1.Binding) ([]webhook.JSONPatchOp, error) {
	if ok, p, err := inject2workload(plugin.TargetContext{
		Binding: &b,
		Values: map[string]interface{}{
			"pvc-name": b.From.Volume.PVCName,
		},
	}, req); ok {
		return p, err
	} else {
		r.Log.Info("unsupported target kind ", "apiVersion", w.APIVersion, "kind", w.Kind)
		return nil, nil
	}
}

func (r *ServiceBindingReconciler) injectSecret(req *admissionv1beta1.AdmissionRequest, w *corev1alpha1.WorkloadReference, b corev1alpha1.Binding) ([]webhook.JSONPatchOp, error) {
	s := b.From.Secret

	secretName := s.Name

	// Read secret name from an object's field
	if f := s.NameFromField; f != nil {
		gv, err := schema.ParseGroupVersion(f.APIVersion)
		if err != nil {
			return nil, err
		}
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gv.Group,
			Version: gv.Version,
			Kind:    f.Kind,
		})
		err = r.Client.Get(context.Background(), client.ObjectKey{
			Namespace: req.Namespace,
			Name:      f.Name,
		}, u)
		if err != nil {
			return nil, err
		}
		arr := strings.Split(f.FieldPath, ".")
		found := false
		if len(arr) > 1 {
			fields := arr[1:]
			secretName, found, err = unstructured.NestedString(u.Object, fields...)
			if err != nil {
				return nil, err
			}
		}
		if !found {
			return nil, fmt.Errorf("fieldPath not found: %s", f.FieldPath)
		}
	}

	if ok, p, err := inject2workload(plugin.TargetContext{
		Binding: &b,
		Values: map[string]interface{}{
			"secret-name": secretName,
		},
	}, req); ok {
		return p, err
	} else {
		r.Log.Info("unsupported target kind ", "apiVersion", w.APIVersion, "kind", w.Kind)
		return nil, nil
	}
}

func inject2workload(pctx plugin.TargetContext, req *admissionv1beta1.AdmissionRequest) (bool, []webhook.JSONPatchOp, error) {
	for _, injector := range plugin.TargetInjectors {
		if !injector.Match(req.Kind) {
			continue
		}

		p, err := injector.Inject(pctx, req.Object)
		if err != nil {
			panic(err)
		}
		return true, p, nil
	}
	return false, nil, nil
}

func healthCheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("service up"))
}
