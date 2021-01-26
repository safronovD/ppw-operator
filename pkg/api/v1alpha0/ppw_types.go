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
	Server     Server     `json:"server,omitempty"`
	Processor  Processor  `json:"processor,omitempty"`
	Controller Controller `json:"controller,omitempty"`
	DataVolume DataVolume `json:"dataVolume,omitempty"`
	Service    Service    `json:"service,omitempty"`
}

type Server struct {
	Name            string `json:"name,omitempty"`
	Label           string `json:"label,omitempty"`
	Size            int32  `json:"size,omitempty"`
	Image           string `json:"image,omitempty"`
	ImagePullSecret string `json:"imagePullSecret,omitempty"`
}

type Processor struct {
	Name            string `json:"name,omitempty"`
	Label           string `json:"label,omitempty"`
	Size            int32  `json:"size,omitempty"`
	VolumeClaimName string `json:"volumeClaimName,omitempty"`
	Image           string `json:"image,omitempty"`
	ImagePullSecret string `json:"imagePullSecret,omitempty"`
	VolumeMountPath string `json:"volumeMountPath,omitempty"`
}

type Controller struct {
	Name            string `json:"name,omitempty"`
	Lifetime        int32 `json:"lifetime,omitempty"`
	VolumeClaimName string `json:"volumeClaimName,omitempty"`
	Image           string `json:"image,omitempty"`
	ImagePullSecret string `json:"imagePullSecret,omitempty"`
	VolumeMountPath string `json:"volumeMountPath,omitempty"`
}

type DataVolume struct {
	Name             string `json:"name,omitempty"`
	Size             string `json:"size,omitempty"`
	StorageClassName string `json:"storageClassName,omitempty"`
}

type Service struct {
	Name            string `json:"name,omitempty"`
	Label           string `json:"label,omitempty"`
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
