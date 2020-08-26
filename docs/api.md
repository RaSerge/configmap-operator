# API

**Note:** This document is generated from code and comments. Do not edit it directly.
## Table of Contents
* [ConfigMapMake](#configmapmake)
* [ConfigMapMakeCondition](#configmapmakecondition)
* [ConfigMapMakeConditionType](#configmapmakeconditiontype)
* [ConfigMapMakeList](#configmapmakelist)
* [ConfigMapMakeSpec](#configmapmakespec)
* [ConfigMapMakeStatus](#configmapmakestatus)
* [ConfigMapTemplate](#configmaptemplate)
* [TemplateMetadata](#templatemetadata)
* [TemplateVariable](#templatevariable)

## ConfigMapMake

ConfigMapMake holds configuration data with embedded configs.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| metadata | Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata | [metav1.ObjectMeta](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta) | false |
| spec | Desired state of the ConfigMapMake. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [ConfigMapMakeSpec](#configmapmakespec) | false |
| status | Observed state of the ConfigMapMake. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status | [ConfigMapMakeStatus](#configmapmakestatus) | false |

[Back to TOC](#table-of-contents)

## ConfigMapMakeCondition

ConfigMapMakeCondition describes the state of a ConfigMapMake.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| type | Type of the condition. | [ConfigMapMakeConditionType](#configmapmakeconditiontype) | true |
| status | Status of the condition: True, False, or Unknown. | [corev1.ConditionStatus](https://pkg.go.dev/k8s.io/api/core/v1#ConditionStatus) | true |
| lastUpdateTime | The last time the condition was updated. | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time) | false |
| lastTransitionTime | Last time the condition transitioned from one status to another. | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time) | false |
| reason | The reason for the last update. | string | false |
| message | A human readable message indicating details about the last update. | string | false |

[Back to TOC](#table-of-contents)

## ConfigMapMakeConditionType

ConfigMapMakeConditionType is a valid value for ConfigMapMakeCondition.Type

| Name | Value | Description |
| ---- | ----- | ----------- |
| ConfigMapMakeRenderFailure | RenderFailure | ConfigMapMakeRenderFailure means that the target config could not be rendered. |

[Back to TOC](#table-of-contents)

## ConfigMapMakeList

ConfigMapMakeList contains a list of ConfigMapMakes.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| metadata | Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#lists-and-simple-kinds | [metav1.ListMeta](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ListMeta) | false |
| items | List of ConfigMapMakes. | [][ConfigMapMake](#configmapmake) | true |

[Back to TOC](#table-of-contents)

## ConfigMapMakeSpec

ConfigMapMakeSpec defines the desired state of a ConfigMapMake.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| template | Template that describes the config that will be rendered. Variable references $(VAR_NAME) in template data are expanded using the ConfigMapMake's variables. If a variable cannot be resolved, the reference in the input data will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. | [ConfigMapTemplate](#configmaptemplate) | false |
| vars | List of template variables. | [][TemplateVariable](#templatevariable) | false |

[Back to TOC](#table-of-contents)

## ConfigMapMakeStatus

ConfigMapMakeStatus describes the observed state of a ConfigMapMake.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| observedGeneration | The generation observed by the ConfigMapMake controller. | int64 | false |
| conditions | Represents the latest available observations of a ConfigMapMake's current state. | [][ConfigMapMakeCondition](#configmapmakecondition) | false |

[Back to TOC](#table-of-contents)

## ConfigMapTemplate

ConfigMapTemplate is a ConfigMap template.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| metadata | Metadata is a stripped down version of the standard object metadata. Its properties will be applied to the metadata of the generated Secret. If no name is provided, the name of the ConfigMapMake will be used. | [TemplateMetadata](#templatemetadata) | false |
| data | Data contains the configuration data. Each key must consist of alphanumeric characters, '-', '_' or '.'. Values with non-UTF-8 byte sequences must use the BinaryData field. The keys stored in Data must not overlap with the keys in the BinaryData field. | map[string]string | false |
| binaryData | BinaryData contains the binary data. Each key must consist of alphanumeric characters, '-', '_' or '.'. BinaryData can contain byte sequences that are not in the UTF-8 range. The keys stored in BinaryData must not overlap with the keys in the Data field. | map[string][]byte | false |

[Back to TOC](#table-of-contents)

## TemplateMetadata

TemplateMetadata is a stripped down version of the standard object metadata.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| name | Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. More info: http://kubernetes.io/docs/user-guide/identifiers#names | string | false |
| labels | Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels | map[string]string | false |
| annotations | Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations | map[string]string | false |

[Back to TOC](#table-of-contents)

## TemplateVariable

TemplateVariable is a template variable.

| Field | Description | Type | Required |
| ----- | ----------- | ---- | -------- |
| name | Name of the template variable. | string | true |
| value | Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the ConfigMapMake. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\". | string | false |
| configValue | SecretValue selects a value by its key in a Secret. | *[corev1.SecretKeySelector](https://pkg.go.dev/k8s.io/api/core/v1#SecretKeySelector) | false |
| configMapValue | ConfigMapValue selects a value by its key in a ConfigMap. | *[corev1.ConfigMapKeySelector](https://pkg.go.dev/k8s.io/api/core/v1#ConfigMapKeySelector) | false |

[Back to TOC](#table-of-contents)
