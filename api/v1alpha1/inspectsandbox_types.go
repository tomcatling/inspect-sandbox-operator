package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen=true

// InspectSandboxSpec defines the desired state of an InspectSandbox
type InspectSandboxSpec struct {
	// Services is a map of service configurations to run in the sandbox
	// +optional
	Services map[string]ServiceSpec `json:"services,omitempty"`

	// AllowDomains is a list of domains that pods can access
	// +optional
	AllowDomains []string `json:"allowDomains,omitempty"`

	// Networks defines logical networks for service communication
	// +optional
	Networks map[string]string `json:"networks,omitempty"`

	// Volumes defines persistent volumes for the sandbox
	// +optional
	Volumes map[string]VolumeSpec `json:"volumes,omitempty"`
}

// +k8s:deepcopy-gen=true

// ServiceSpec defines a service to be run in the sandbox
type ServiceSpec struct {
	// Image is the container image to use
	Image string `json:"image"`

	// RuntimeClassName specifies the container runtime to use (e.g., gvisor)
	// +optional
	RuntimeClassName string `json:"runtimeClassName,omitempty"`

	// Command to run in the container
	// +optional
	Command []string `json:"command,omitempty"`

	// Args for the command
	// +optional
	Args []string `json:"args,omitempty"`

	// WorkingDir for the container
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// DNSRecord indicates whether to create a DNS record for this service
	// +optional
	DNSRecord bool `json:"dnsRecord,omitempty"`

	// AdditionalDNSRecords provides additional domain names for this service
	// +optional
	AdditionalDNSRecords []string `json:"additionalDnsRecords,omitempty"`

	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Volumes to mount
	// +optional
	Volumes []string `json:"volumes,omitempty"`

	// Resources specifies resource requirements
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Networks this service belongs to
	// +optional
	Networks []string `json:"networks,omitempty"`
}

// +k8s:deepcopy-gen=true

// VolumeSpec defines a persistent volume for the sandbox
type VolumeSpec struct {
	// Size of the volume
	Size string `json:"size,omitempty"`

	// StorageClass to use
	// +optional
	StorageClass string `json:"storageClass,omitempty"`
}

// +k8s:deepcopy-gen=true

// InspectSandboxStatus defines the observed state of InspectSandbox
type InspectSandboxStatus struct {
	// Conditions represent the latest available observations of the sandbox state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Services represents the status of individual services
	// +optional
	Services map[string]ServiceStatus `json:"services,omitempty"`
}

// +k8s:deepcopy-gen=true

// ServiceStatus provides status information for a service
type ServiceStatus struct {
	// Ready indicates whether the service is ready
	Ready bool `json:"ready"`

	// Message provides additional status information
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=isbox

// InspectSandbox is the Schema for the inspectsandboxes API
type InspectSandbox struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InspectSandboxSpec   `json:"spec,omitempty"`
	Status InspectSandboxStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InspectSandboxList contains a list of InspectSandbox resources
type InspectSandboxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InspectSandbox `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InspectSandbox{}, &InspectSandboxList{})
}
