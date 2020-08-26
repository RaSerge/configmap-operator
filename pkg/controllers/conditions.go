package controllers

import (
	"github.com/raserge/configmap-operator/pkg/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CreateVariablesErrorReason is the reason given when required ConfigMapMake
	// variables cannot be resolved.
	CreateVariablesErrorReason = "CreateVariablesError"

	internalError = "InternalError"
)

// NewConfigMapMakeCondition creates a new deployment condition.
func NewConfigMapMakeCondition(typ v1alpha1.ConfigMapMakeConditionType, status corev1.ConditionStatus, reason, message string) *v1alpha1.ConfigMapMakeCondition {
	return &v1alpha1.ConfigMapMakeCondition{
		Type:               typ,
		Status:             status,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetConfigMapMakeCondition returns the condition with the provided type.
func GetConfigMapMakeCondition(status v1alpha1.ConfigMapMakeStatus, typ v1alpha1.ConfigMapMakeConditionType) *v1alpha1.ConfigMapMakeCondition {
	for _, c := range status.Conditions {
		if c.Type == typ {
			return &c
		}
	}
	return nil
}

// SetConfigMapMakeCondition updates the status to include the provided condition.
// If the condition already exists with the same status, reason, and message then it is not updated.
func SetConfigMapMakeCondition(status *v1alpha1.ConfigMapMakeStatus, cond v1alpha1.ConfigMapMakeCondition) {
	if prev := GetConfigMapMakeCondition(*status, cond.Type); prev != nil {
		if prev.Status == cond.Status &&
			prev.Reason == cond.Reason &&
			prev.Message == cond.Message {
			return
		}
		// Do not update lastTransitionTime if the status of the condition doesn't change.
		if prev.Status == cond.Status {
			cond.LastTransitionTime = prev.LastTransitionTime
		}
	}
	RemoveConfigMapMakeCondition(status, cond.Type)
	status.Conditions = append(status.Conditions, cond)
}

// RemoveConfigMapMakeCondition removes the condition with the provided type.
func RemoveConfigMapMakeCondition(status *v1alpha1.ConfigMapMakeStatus, typ v1alpha1.ConfigMapMakeConditionType) {
	var conds []v1alpha1.ConfigMapMakeCondition
	for _, c := range status.Conditions {
		if c.Type == typ {
			continue
		}
		conds = append(conds, c)
	}
	status.Conditions = conds
}
