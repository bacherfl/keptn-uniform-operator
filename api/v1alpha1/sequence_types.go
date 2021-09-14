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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SequenceScope struct {
	Project string `json:"project"`
	Stage   string `json:"stage"`
	Service string `json:"service"`
}

type TaskDefinition struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

type TaskState struct {
	TriggeredID string `json:"triggered_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
}

type TaskResult struct {
	Name             string                 `json:"name"`
	Result           string                 `json:"result"`
	Status           string                 `json:"status"`
	ResultProperties map[string]interface{} `json:"result_properties"`
}

// SequenceSpec defines the desired state of Sequence
type SequenceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Sequence. Edit sequence_types.go to remove/update
	Name  string           `json:"name,omitempty"`
	Scope SequenceScope    `json:"scope"`
	Tasks []TaskDefinition `json:"tasks"`
}

// SequenceStatus defines the observed state of Sequence
type SequenceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State          string       `json:"state"`
	KeptnContextID string       `json:"keptn_context_id"`
	PreviousTasks  []TaskResult `json:"previous_tasks"`
	CurrentTask    TaskState    `json:"current_task"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Sequence is the Schema for the sequences API
type Sequence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SequenceSpec   `json:"spec,omitempty"`
	Status SequenceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SequenceList contains a list of Sequence
type SequenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sequence `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sequence{}, &SequenceList{})
}
