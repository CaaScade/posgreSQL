package controller

import (
	"strings"

	"github.com/caascade/posgreSQL/posgresql/client"
	"github.com/caascade/posgreSQL/posgresql/resource"

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/Sirupsen/logrus"
)

func handleUpdate(oldObj, newObj *resource.Application) {
	if (oldObj.Status.State == "Configured" && newObj.Status.State == "Configured") && (oldObj.Spec.Scale != newObj.Spec.Scale) {
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
	kClient := client.GetClient()
	dep, err := kClient.ExtensionsV1beta1().Deployments(apiv1.NamespaceDefault).Get("slave", metav1.GetOptions{})
	if err != nil {
		log.Errorf("Error getting slave deployment %s", err.Error())
		return
	}
	if newObj.Spec.Scale > 8 {
		log.Errorf("cannot update scale to a value greater than 8")
		return
	}
	if int32(newObj.Spec.Scale) != *dep.Spec.Replicas {
		val := int32(newObj.Spec.Scale)
		dep.Spec.Replicas = &val
	}
	updatedDep, err := kClient.ExtensionsV1beta1().Deployments(apiv1.NamespaceDefault).Update(dep)
	if err != nil {
		log.Errorf("Error updating the scale of deployment %s", err.Error())
	}
	if *updatedDep.Spec.Replicas != int32(newObj.Spec.Scale) {
		log.Errorf("Scale of the deployment has not been updated!")
	}
}

func updateDeployment(oldObj, newObj *resource.Application) {

}

func updateState(oldObj, newObj *resource.Application) {
	if newObj.Status.State == "Recovery" {
		if oldObj.Status.State == "Recovery" {
			return
		}
		kClient := client.GetClient()
		svc, err := kClient.CoreV1().Services(apiv1.NamespaceDefault).Get("posgres", metav1.GetOptions{})
		if err != nil {
			log.Errorf("XXXXXXXXXXXX Error updating service posgres %s XXXXXXXXX", err.Error())
			return
		}
		slaveName := ""
		for _, x := range newObj.Status.Addresses.Slaves {
			slaveName = strings.Replace(x.IP, ".", "-", -1)
		}
		if slaveName == "" {
			log.Errorf("XXXXNo slave left for recoveryXXXXX")
			return
		}
		if _, ok := svc.Spec.Selector[slaveName]; ok {
			return
		}
		svc.Spec.Selector = map[string]string{}
		svc.Spec.Selector[slaveName] = "true"
		_, err = kClient.CoreV1().Services(apiv1.NamespaceDefault).Update(svc)
		if err != nil {
			log.Errorf("XXXXX error recoverinng XXXXXX", err)
		}
		return
	}
	if oldObj.Status.State == "Created" || oldObj.Status.State == "Restore" {
		if newObj.Status.State == "Configured" {
			shouldRestore := false
			if oldObj.Status.State == "Restore" {
				shouldRestore = true
			}
			//Ensure pre-requisites are available
			shouldProtect := ensureSecret()
			shouldStore := ensureStorage()
			//Deploy pods
			deployPods(newObj, shouldProtect, shouldStore, shouldRestore)
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
func ensureStorage() bool {
	quantityStr := "10G"
	quantity, err := apiResource.ParseQuantity(quantityStr)
	if err != nil {
		log.Errorf("Error parsing quantity %s", err.Error())
		return false
	}
	pvc := apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "posgres-pvc",
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			Resources: apiv1.ResourceRequirements{
				Requests: map[apiv1.ResourceName]apiResource.Quantity{
					apiv1.ResourceStorage: quantity,
				},
			},
		},
	}

	kClient := client.GetClient()

	_, err = kClient.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault).Create(&pvc)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return true
		}
		log.Errorf("Error creating pvc %s", err.Error())
		return false
	}
	return true
}

//deploy pods with values in newObj
func deployPods(newObj *resource.Application, passwdProtected, shouldPersist, shouldRestore bool) {
	runAsUser := int64(999)
	var masterVolSrc apiv1.VolumeSource
	if shouldPersist {
		masterVolSrc = apiv1.VolumeSource{
			PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
				ClaimName: "posgres-pvc",
			},
		}
	} else {
		masterVolSrc = apiv1.VolumeSource{
			EmptyDir: &apiv1.EmptyDirVolumeSource{},
		}
	}
	masterArgs := []string{
		"--init-master",
		"--controller-address",
		controllerIP,
	}
	if shouldRestore {
		masterArgs = append(masterArgs, "--restore")
	}
	masterEnv := []apiv1.EnvVar{
		{
			Name: "SELF_IP",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	}
	if shouldRestore {
		masterEnv = append(
			masterEnv,
			apiv1.EnvVar{
				Name:  "PUBLIC_KEY",
				Value: newObj.Spec.PublicKey,
			},
			apiv1.EnvVar{
				Name:  "SECRET_KEY",
				Value: newObj.Spec.SecretKey,
			})
	}

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
					Name:            "posgres-init",
					Image:           "wlan0/posgres-sidecar:v0.0.1",
					Args:            masterArgs,
					Env:             masterEnv,
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
				{
					Name:  "sidecar",
					Image: "wlan0/posgres-sidecar:v0.0.1",
					Args: []string{
						"--sidecar",
						"--controller-address",
						controllerIP,
						"--sidecar-type",
						"master",
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
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "data-dir",
							MountPath: "/var/lib/postgresql/data/",
						},
					},
				},
			},
			NodeSelector: map[string]string{
				"name": "posgres-master",
			},
			Volumes: []apiv1.Volume{
				{
					Name:         "data-dir",
					VolumeSource: masterVolSrc,
				},
			},
		},
	}

	slaveEnv := []apiv1.EnvVar{
		{
			Name: "SELF_IP",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		apiv1.EnvVar{
			Name:  "PUBLIC_KEY",
			Value: newObj.Spec.PublicKey,
		},
		apiv1.EnvVar{
			Name:  "SECRET_KEY",
			Value: newObj.Spec.SecretKey,
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
					Command: []string{
						"postgres",
					},
					Args: []string{
						"-D",
						"/var/lib/postgresql/data/",
					},
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "data-dir",
							MountPath: "/var/lib/postgresql/data/",
						},
					},
				},
				{
					Name:  "sidecar",
					Image: "wlan0/posgres-sidecar:v0.0.1",
					Args: []string{
						"--sidecar",
						"--controller-address",
						controllerIP,
						"--sidecar-type",
						"slave",
					},
					Env: slaveEnv,
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
			Name: "posgres",
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
	kClient := client.GetClient()
	_, err := kClient.CoreV1().Services(apiv1.NamespaceDefault).Create(&posgresMasterService)
	if err != nil {
		log.Errorf("Error creating master service %v", err)
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
