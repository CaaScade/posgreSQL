package app

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/caascade/posgreSQL/posgresql/client"
	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type _ interface {
	GetApp() resource.Application
	UpdateApp(resource.Application) (int, string)
	SetPassword(resource.Password) (int, string)
	GetAddresses(resource.Addresses) (int, string)
}

const task = "create-posgres-app"

var registration_uuid string

func Init(uuid string) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		log.Info("initializing task: create-posgres-app")
		createApp()
		executor.ReturnToken(task, uuid)
	}()
}

func createApp() {
	client, _ := resource.GetApplicationClientScheme()
	posgresApp := &resource.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "posgres",
		},
		Spec: resource.ApplicationSpec{
			Scale: 0,
		},
		Status: resource.ApplicationStatus{
			State: "Created",
		},
	}
	var result resource.Application
	err := client.Post().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Body(posgresApp).
		Do().Into(&result)

	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return
		}
		executor.SetErrorState(registration_uuid, err)
	}
}

func GetApp() resource.Application {
	client, _ := resource.GetApplicationClientScheme()
	var posgresApp resource.Application
	err := client.Get().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Do().Into(&posgresApp)
	if err != nil {
		executor.SetErrorState(registration_uuid, err)
	}
	return posgresApp
}

func UpdateApp(posgresApp []byte) (int, string) {
	client, _ := resource.GetApplicationClientScheme()
	newAppObj := resource.Application{}
	err := client.Put().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Body(posgresApp).
		Do().Into(&newAppObj)
	if err != nil {
		return 500, err.Error()
	}
	newAppObjBytes, err := json.Marshal(newAppObj)
	if err != nil {
		return 500, err.Error()
	}
	return 200, string(newAppObjBytes)
}

func SetPassword(passwd resource.Password) (int, string) {
	kClient := client.GetClient()
	secret := apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "posgres-secret",
		},
		StringData: map[string]string{
			"postgres-password": passwd.Password,
		},
	}
	secretObj, err := kClient.CoreV1().Secrets(apiv1.NamespaceDefault).Create(&secret)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return 500, err.Error()
		}
	}
	secretData, err := json.Marshal(secretObj)
	if err != nil {
		return 500, err.Error()
	}
	return 200, string(secretData)
}

func GetAddresses() (int, string) {
	kClient := client.GetClient()
	masterService, err := kClient.CoreV1().Services(apiv1.NamespaceDefault).Get("posgres-master", metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return 500, err.Error()
		}
	}
	slaveService, err := kClient.CoreV1().Services(apiv1.NamespaceDefault).Get("posgres-slave", metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return 500, err.Error()
		}
	}

	masterPort := 0
	if len(masterService.Spec.Ports) > 0 {
		masterPort = int(masterService.Spec.Ports[0].Port)
	}

	slavePort := 0
	if len(slaveService.Spec.Ports) > 0 {
		slavePort = int(slaveService.Spec.Ports[0].Port)
	}

	addrs := resource.Addresses{
		MasterIP:   masterService.Spec.ClusterIP,
		MasterPort: masterPort,

		SlaveIP:   slaveService.Spec.ClusterIP,
		SlavePort: slavePort,
	}

	addresses, err := json.Marshal(addrs)
	if err != nil {
		return 500, err.Error()
	}
	return 200, string(addresses)
}
