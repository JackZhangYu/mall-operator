/*
Copyright 2022.

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

package v1

import (
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MallWebSpec defines the desired state of MallWeb
type MallWebSpec struct {
	Image         string `json:"image"`
	Port          *int32 `json:"port"`
	SinglePodsQPS *int32 `json:"singlePodsQPS"`
	TotalQPS      *int32 `json:"totalQPS,omitempty"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
}

// MallWebStatus defines the observed state of MallWeb
type MallWebStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	RealQPS *int32 `json:"realQPS"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image",description="The Docker image of etcd"
//+kubebuilder:printcolumn:name="Port",type="integer",priority=1,JSONPath=".spec.port",description="container port"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:subresource:status

// MallWeb is the Schema for the mallwebs API
type MallWeb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MallWebSpec   `json:"spec,omitempty"`
	Status MallWebStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MallWebList contains a list of MallWeb
type MallWebList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MallWeb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MallWeb{}, &MallWebList{})
}

func (e *MallWeb) String() string {
	var realQPS string
	if nil == e.Status.RealQPS {
		realQPS = ""
	} else {
		realQPS = strconv.Itoa(int(*e.Status.RealQPS))
	}

	return fmt.Sprintf("Image [%s], Port [%d], SinglePodQPS [%d],TotalQPS[%d], ReadQPS [%s]",
		e.Spec.Image,
		*e.Spec.Port,
		*e.Spec.SinglePodsQPS,
		*e.Spec.TotalQPS,
		realQPS)
}
