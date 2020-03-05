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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceBindingSpec defines the desired state of ServiceBinding
type ServiceBindingSpec struct {
	// Important: Run "make generate manifests" to regenerate code after modifying

	Bindings []Binding `json:"bindings,omitempty"`

	WorkloadRef *WorkloadReference `json:"workloadRef,omitempty"`
}

type Binding struct {
	// Source indicates the source object to get binding data from.
	From DataSource `json:"from,omitempty"`

	// Target indicates the target objects to inject the binding data to.
	To DataTarget `json:"to,omitempty"`
}

type DataSource struct {
	Secret *SecretSource `json:"secret,omitempty"`
}
type SecretSource struct {
	SecretName     string           `json:"secretName,omitempty"`
	SecretNameFrom SecretNameSource `json:"secretNameFrom,omitempty"`
}

type SecretNameSource struct {
	// APIVersion of the referenced workload.
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referenced workload.
	Kind string `json:"kind,omitempty"`

	// Name of the referenced workload.
	Name string `json:"name,omitempty"`

	// Namespace of the referenced workload.
	Namespace string `json:"namespace,omitempty"`

	// The path of the field whose value is the secret name. E.g. ".status.output-secret" .
	FieldPath string `json:"fieldPath,omitempty"`
}

// Target defines what target objects to inject the binding data to.
type DataTarget struct {
	// The path of the file where the data source is mounted.
	FilePath string `json:"filePath,omitempty"`

	// Env indicates whether to inject all `K=V` pairs from data source into environment variables.
	Env bool `json:"env,omitempty"`
}

// A WorkloadReference refers to an OAM workload resource.
type WorkloadReference struct {
	// APIVersion of the referenced workload.
	APIVersion string `json:"apiVersion"`

	// Kind of the referenced workload.
	Kind string `json:"kind"`

	// Name of the referenced workload.
	Name string `json:"name"`
}

type ServiceBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// ServiceBinding is the Schema for the servicebindings API
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec,omitempty"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceBindingList contains a list of ServiceBinding
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBinding{}, &ServiceBindingList{})
}
