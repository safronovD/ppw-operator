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

package v1alpha0

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PpwSpec defines the desired state of Ppw
type PpwSpec struct {
	Processor struct {
		Size    int32  `json:"size"`
		PvcName string `json:"pvcName"`
		Image   string `json:"image"`
	} `json:"processor"`

	Server struct {
		Size  int32  `json:"size"`
		Image string `json:"image"`
	} `json:"server"`

	mlController struct {
		PvcName string `json:"pvcName"`
		Image   string `json:"image"`
	}
}

// PpwStatus defines the observed state of Ppw
type PpwStatus struct {
	Nodes []string `json:"nodes"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Ppw is the Schema for the ppws API
type Ppw struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PpwSpec   `json:"spec,omitempty"`
	Status PpwStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PpwList contains a list of Ppw
type PpwList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ppw `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ppw{}, &PpwList{})
}
