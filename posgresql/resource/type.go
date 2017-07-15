package resource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const AppResourcePlural = "applications"
const AppResource = "application"
const AppResourceGroup = "appextensions.k8s.io"

type Password struct {
	Password string `json:"password"`
}

type Addresses struct {
	MasterIP   string `json:"masterIP"`
	MasterPort int    `json:"masterPort"`

	SlaveIP   string `json:"slaveIP"`
	SlavePort int    `json:"slavePort"`
}

type Application struct {
	metav1.TypeMeta   `json:", inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ApplicationSpec   `json:"spec"`
	Status            ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {
	Scale             int                  `json:"scale"`
	DeploymentType    DeploymentType       `json:"deploymentType"`
	SecretRef         ApplicationSecretRef `json:"secretRef"`
	ResourceNamespace string               `json:"resourceNamespace"`
}

type DeploymentType string

const (
	DeploymentTypeIsolation DeploymentType = "isolation"
	DeploymentTypeScaling   DeploymentType = "scaling"
)

type ApplicationSecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type ApplicationStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Application `json:"items"`
}
