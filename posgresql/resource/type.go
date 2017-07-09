package resource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const AppResourcePlural = "applications"
const AppResource = "application"
const AppResourceGroup = "appextensions.k8s.io"

type Application struct {
	metav1.TypeMeta   `json:", inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ApplicationSpec   `json:"spec"`
	Status            ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {
	Scale          int            `json:"scale"`
	DeploymentType DeploymentType `json:"deploymentType"`
}

type DeploymentType string

const (
	DeploymentTypeIsolation DeploymentType = "isolation"
	DeploymentTypeScaling   DeploymentType = "scaling"
)

type ApplicationStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Application `json:"items"`
}
