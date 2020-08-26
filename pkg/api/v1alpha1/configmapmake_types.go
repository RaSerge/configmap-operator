package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&ConfigMapMake{}, &ConfigMapMakeList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ConfigMapMake holds configuration data.
type ConfigMapMake struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Desired state of the ConfigMapMake.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec ConfigMapMakeSpec `json:"spec,omitempty"`

	// Observed state of the ConfigMapMake.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status ConfigMapMakeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigMapMakeList contains a list of ConfigMapMakes.
type ConfigMapMakeList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#lists-and-simple-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of ConfigMapMakes.
	Items []ConfigMapMake `json:"items"`
}

// ConfigMapMakeSpec defines the desired state of a ConfigMapMake.
type ConfigMapMakeSpec struct {
	// Template that describes the config that will be rendered.
	// Variable references $(VAR_NAME) in template data are expanded using the
	// ConfigMapMake's variables. If a variable cannot be resolved, the reference
	// in the input data will be unchanged. The $(VAR_NAME) syntax can be escaped
	// with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded,
	// regardless of whether the variable exists or not.
	// +optional
	Template ConfigMapTemplate `json:"template,omitempty"`

	// List of template variables.
	// +optional
	Vars []TemplateVariable `json:"vars,omitempty"`
}

// ConfigMapTemplate is a ConfigMap template.
type ConfigMapTemplate struct {
	// Metadata is a stripped down version of the standard object metadata.
	// Its properties will be applied to the metadata of the generated Secret.
	// If no name is provided, the name of the ConfigMapMake will be used.
	// +optional
	Metadata TemplateMetadata `json:"metadata,omitempty"`

	// Data contains the configuration data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// Values with non-UTF-8 byte sequences must use the BinaryData field.
	// The keys stored in Data must not overlap with the keys in
	// the BinaryData field.
	// +optional
	Data map[string]string `json:"data,omitempty"`

	// BinaryData contains the binary data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// BinaryData can contain byte sequences that are not in the UTF-8 range.
	// The keys stored in BinaryData must not overlap with the keys in
	// the Data field.
	// +optional
	BinaryData map[string][]byte `json:"binaryData,omitempty"`
}

// TemplateMetadata is a stripped down version of the standard object metadata.
type TemplateMetadata struct {
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// TemplateVariable is a template variable.
type TemplateVariable struct {
	// Name of the template variable.
	Name string `json:"name"`

	// Variable references $(VAR_NAME) are expanded using the previous defined
	// environment variables in the ConfigMapMake. If a variable cannot be resolved,
	// the reference in the input string will be unchanged. The $(VAR_NAME) syntax
	// can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will
	// never be expanded, regardless of whether the variable exists or not.
	// Defaults to "".
	// +optional
	Value string `json:"value,omitempty"`
	// SecretValue selects a value by its key in a Secret.
	// +optional
	SecretValue *corev1.SecretKeySelector `json:"configValue,omitempty"`
	// ConfigMapValue selects a value by its key in a ConfigMap.
	// +optional
	ConfigMapValue *corev1.ConfigMapKeySelector `json:"configMapValue,omitempty"`
}

// ConfigMapMakeStatus describes the observed state of a ConfigMapMake.
type ConfigMapMakeStatus struct {
	// The generation observed by the ConfigMapMake controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a ConfigMapMake's current state.
	// +listType=map
	// +listMapKey=type
	// +listMapKeys=type
	// +optional
	Conditions []ConfigMapMakeCondition `json:"conditions,omitempty"`
}

// ConfigMapMakeCondition describes the state of a ConfigMapMake.
type ConfigMapMakeCondition struct {
	// Type of the condition.
	Type ConfigMapMakeConditionType `json:"type"`
	// Status of the condition: True, False, or Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time the condition was updated.
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the last update.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the last update.
	// +optional
	Message string `json:"message,omitempty"`
}

// ConfigMapMakeConditionType is a valid value for ConfigMapMakeCondition.Type
type ConfigMapMakeConditionType string

const (
	// ConfigMapMakeRenderFailure means that the target config could not be
	// rendered.
	ConfigMapMakeRenderFailure ConfigMapMakeConditionType = "RenderFailure"
)
