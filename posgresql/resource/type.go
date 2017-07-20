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
	Master Address   `master`
	Slaves []Address `slaves`
}

type Address struct {
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	LastUpdated int64  `json:"lastUpdated"`
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
	State     string    `json:"state,omitempty"`
	Message   string    `json:"message,omitempty"`
	Addresses Addresses `json:"addresses,omitempty"`
}

type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Application `json:"items"`
}
