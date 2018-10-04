package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FalcoRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []FalcoRule `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FalcoRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FalcoRuleSpec   `json:"spec"`
	Status            FalcoRuleStatus `json:"status,omitempty"`
}

type FalcoRuleSpec struct {
	Rule string `json:"rule"`
	Desc string `json:"desc"`
	Condition string `json:"condition"`
	Output string `json:"output"`
	priority string `json:"priority"`
}

type FalcoRuleStatus struct {
	// Fill me
}
