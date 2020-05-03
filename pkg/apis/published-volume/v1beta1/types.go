package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PublishedVolume is a top-level type
type PublishedVolume struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PublishedVolumeSpec `json:"spec"`
}

type PublishedVolumeSpec struct {
	Node         string                 `json:"node"`
	TargetPath   string                 `json:"targetPath"`
	VolumeHandle string                 `json:"volumeHandle"`
	Options      map[string]string      `json:"options"`
	Pod          PublishedVolumeSpecPod `json:"pod"`
}

type PublishedVolumeSpecPod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PublishedVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []PublishedVolume `json:"items"`
}
