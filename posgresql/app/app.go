package app

import (
	"encoding/json"
	"time"

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
	UpdateAddresses(resource.Addresses) (int, string)
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
	appl := GetApp()

	addresses, err := json.Marshal(appl.Status.Addresses)
	if err != nil {
		return 500, err.Error()
	}
	return 200, string(addresses)
}

func UpdateAddresses(addrs resource.Addresses) (int, string) {
	count := 0
	for {
		count++
		if count == 10 {
			break
		}
		client, _ := resource.GetApplicationClientScheme()
		appl := GetApp()
		appl.Status.Addresses = addrs
		appData, err := json.Marshal(appl)
		if err != nil {
			return 500, err.Error()
		}
		newAppObj := resource.Application{}
		err = client.Put().
			Resource(resource.AppResourcePlural).
			Namespace(apiv1.NamespaceDefault).
			Name("posgres").
			Body(appData).
			Do().Into(&newAppObj)
		if err == nil {
			data, _ := json.Marshal(newAppObj.Status.Addresses)
			return 200, string(data)
		}
		time.Sleep(1 * time.Second)
	}
	return 500, "Failed to update addresses after 10 retries"
}

func equivalentAddrs(left, right resource.Addresses) bool {
	if left.MasterIP == right.MasterIP {
		if left.MasterPort == right.MasterPort {
			if left.SlaveIP == right.SlaveIP {
				if left.SlavePort == right.SlavePort {
					return true
				}
			}
		}
	}
	return false
}
