package controller

import (
	"github.com/caascade/posgreSQL/posgresql/client"
	"github.com/caascade/posgreSQL/posgresql/resource"

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/Sirupsen/logrus"
)

func handleUpdate(oldObj, newObj *resource.Application) {
	if oldObj.Spec.Scale != newObj.Spec.Scale {
		log.Infof("scaling posgres from %d to %d", oldObj.Spec.Scale, newObj.Spec.Scale)
		updateScale(oldObj, newObj)
	}
	if oldObj.Spec.DeploymentType != newObj.Spec.DeploymentType {
		log.Infof("updating posgres from %s to %s", oldObj.Spec.DeploymentType, newObj.Spec.DeploymentType)
		updateDeployment(oldObj, newObj)
	}
	if oldObj.Status.State != newObj.Status.State {
		log.Infof("updating state from %s to %s", oldObj.Status.State, newObj.Status.State)
		updateState(oldObj, newObj)
	}
}

func updateScale(oldObj, newObj *resource.Application) {

}

func updateDeployment(oldObj, newObj *resource.Application) {

}

func updateState(oldObj, newObj *resource.Application) {
	if oldObj.Status.State == "Created" {
		if newObj.Status.State == "Configured" {
			//Ensure pre-requisites are available
			shouldProtect := ensureSecret()
			ensureStorage()
			//Deploy pods
			deployPods(newObj, shouldProtect)
			//Create service
			createService()
		}
	}
}

// configure secret with default values
func ensureSecret() bool {
	kClient := client.GetClient()
	_, err := kClient.CoreV1().Secrets(apiv1.NamespaceDefault).Get("posgres-secret", metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

//configure storage with default values
func ensureStorage() {}

//deploy pods with values in newObj
func deployPods(newObj *resource.Application, passwdProtected bool) {
	runAsUser := int64(999)
	posgresMasterPodTemplate := apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgres-master",
			Labels: map[string]string{
				"name": "posgres-master",
			},
		},
		Spec: apiv1.PodSpec{
			InitContainers: []apiv1.Container{
				{
					Name:  "posgres-init",
					Image: "wlan0/posgres-sidecar:v0.0.1",
					Args: []string{
						"--init-master",
						"--controller-address",
						controllerIP,
					},
					Env: []apiv1.EnvVar{
						{
							Name: "SELF_IP",
							ValueFrom: &apiv1.EnvVarSource{
								FieldRef: &apiv1.ObjectFieldSelector{
									FieldPath: "status.podIP",
								},
							},
						},
					},
					ImagePullPolicy: apiv1.PullAlways,
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "data-dir",
							MountPath: "/var/lib/postgresql/data/",
						},
					},
				},
			},
			Containers: []apiv1.Container{
				{
					Name:  "posgres-master",
					Image: "postgres:9.6.2",
					SecurityContext: &apiv1.SecurityContext{
						RunAsUser: &runAsUser,
					},
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "data-dir",
							MountPath: "/var/lib/postgresql/data/",
						},
					},
					Command: []string{
						"postgres",
					},
					Args: []string{
						"-D",
						"/var/lib/postgresql/data/",
					},
				},
			},
			NodeSelector: map[string]string{
				"name": "posgres-master",
			},
			Volumes: []apiv1.Volume{
				{
					Name: "data-dir",
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	posgresSlavePodTemplate := apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: "posgres-slave",
			Labels: map[string]string{
				"name": "posgres-slave",
			},
		},
		Spec: apiv1.PodSpec{
			InitContainers: []apiv1.Container{
				{
					Name:  "posgres-init",
					Image: "wlan0/posgres-sidecar:v0.0.1",
					Args: []string{
						"--init-slave",
						"--controller-address",
						controllerIP,
					},
					Env: []apiv1.EnvVar{
						{
							Name: "SELF_IP",
							ValueFrom: &apiv1.EnvVarSource{
								FieldRef: &apiv1.ObjectFieldSelector{
									FieldPath: "status.podIP",
								},
							},
						},
					},
					ImagePullPolicy: apiv1.PullAlways,
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "data-dir",
							MountPath: "/var/lib/postgresql/data/",
						},
					},
				},
			},
			Containers: []apiv1.Container{
				{
					Name:  "posgres-slave",
					Image: "postgres:9.6.2",
					SecurityContext: &apiv1.SecurityContext{
						RunAsUser: &runAsUser,
					},
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "data-dir",
							MountPath: "/var/lib/postgresql/data/",
						},
					},
				},
			},
			NodeSelector: map[string]string{
				"name": "posgres-slave",
			},
			Volumes: []apiv1.Volume{
				{
					Name: "data-dir",
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
	posgresMasterDeployment := extensionsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "master",
			Labels: map[string]string{
				"name": "posgres-master",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Template: posgresMasterPodTemplate,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "posgres-master",
				},
			},
		},
	}
	posgresSlaveDeployment := extensionsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "slave",
			Labels: map[string]string{
				"name": "posgres-slave",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Template: posgresSlavePodTemplate,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "posgres-slave",
				},
			},
		},
	}
	posgresMasterService := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "posgres-master",
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"name": "posgres-master",
			},
			Ports: []apiv1.ServicePort{
				{
					Port: 5432,
				},
			},
		},
	}
	posgresSlaveService := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "posgres-slave",
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"name": "posgres-slave",
			},
			Ports: []apiv1.ServicePort{
				{
					Port: 5432,
				},
			},
		},
	}
	kClient := client.GetClient()
	_, err := kClient.CoreV1().Services(apiv1.NamespaceDefault).Create(&posgresMasterService)
	if err != nil {
		log.Errorf("Error creating master service %v", err)
		return
	}
	_, err = kClient.CoreV1().Services(apiv1.NamespaceDefault).Create(&posgresSlaveService)
	if err != nil {
		log.Errorf("Error creating slave service %v", err)
		return
	}
	_, err = kClient.ExtensionsV1beta1().Deployments(apiv1.NamespaceDefault).Create(&posgresMasterDeployment)
	if err != nil {
		log.Errorf("Error creating master deployment %v", err)
		return
	}
	_, err = kClient.ExtensionsV1beta1().Deployments(apiv1.NamespaceDefault).Create(&posgresSlaveDeployment)
	if err != nil {
		log.Errorf("Error creating slave deployment %v", err)
		return
	}
}

//create service with default values
func createService() {}
