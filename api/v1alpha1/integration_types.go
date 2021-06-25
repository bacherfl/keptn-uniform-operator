/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SecretRef struct {
	SecretName string `json:"secretName"`
	Field      string `json:"field"`
}

type APITokenSpec struct {
	Value string `json:"value,omitempty"`
	// Source for the environment variable's value. Cannot be used if value is not empty.
	// +optional
	FromSecret *SecretRef `json:"fromSecret,omitempty"`
}

type RemoteExecutionPlaneSpec struct {
	APIURL   string       `json:"apiURL"`
	APIToken APITokenSpec `json:"apiToken"`
	// +optional
	DisableSSLVerification *bool `json:"disableSSLVerification"`
}

// IntegrationSpec defines the desired state of Integration
type IntegrationSpec struct {
	Name   string   `json:"name"`
	Image  string   `json:"image"`
	Events []string `json:"events"`
	// Env allows to pass environment variables to the integration
	Env []corev1.EnvVar `json:"env,omitempty"`
	//Remote configuration for the remote execution plane mode
	// +optional
	Remote *RemoteExecutionPlaneSpec `json:"remote"`
}

// IntegrationStatus defines the observed state of Integration
type IntegrationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Integration is the Schema for the integrations API
type Integration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationSpec   `json:"spec,omitempty"`
	Status IntegrationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IntegrationList contains a list of Integration
type IntegrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Integration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Integration{}, &IntegrationList{})
}
