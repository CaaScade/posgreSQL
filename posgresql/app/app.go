package app

import (
	"encoding/json"
	"fmt"
	"sync"
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
	ScaleApp(scale int) (int, string)

	SetPassword(resource.Password) (int, string)
	GetAddresses(resource.Addresses) (int, string)
	UpdateAddresses(resource.Address, string) (int, string)
	ResetSlaves() (int, string)
}

const task = "create-posgres-app"

var registration_uuid string
var updateLock sync.Mutex

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
		TypeMeta: metav1.TypeMeta{
			Kind: "application",
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
		log.Infof("Error creating app %s", err.Error())
		UpdateState("Created")
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
	updateLock.Lock()
	defer updateLock.Unlock()

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
	updateLock.Lock()
	defer updateLock.Unlock()

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

func UpdateAddresses(addr resource.Address, addrType string) (int, string) {
	updateLock.Lock()
	defer updateLock.Unlock()

	addr.LastUpdated = time.Now().Unix()

	if addrType == "master" {
		return updateMasterAddress(addr)
	} else if addrType == "slave" {
		return addSlaveAddress(addr)
	}
	return 500, fmt.Sprintf("Unknown addr type %s", addrType)
}

func DeleteSlaveAddress(addr resource.Address) (int, string) {
	client, _ := resource.GetApplicationClientScheme()
	var posgresApp resource.Application
	err := client.Get().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Do().Into(&posgresApp)
	if err != nil {
		return 500, fmt.Sprintf("Error getting app obj %s", err.Error())
	}
	newAppObj := resource.Application{}
	if len(posgresApp.Status.Addresses.Slaves) == 0 {
		return 200, "Success"
	}

	newSlaveAddrs := []resource.Address{}

	for i := range posgresApp.Status.Addresses.Slaves {
		if posgresApp.Status.Addresses.Slaves[i].IP == addr.IP {
			if posgresApp.Status.Addresses.Slaves[i].Port == addr.Port {
				continue
			}
		}
		newSlaveAddrs = append(newSlaveAddrs, posgresApp.Status.Addresses.Slaves[i])
	}

	posgresApp.Status.Addresses.Slaves = newSlaveAddrs

	data, err := json.Marshal(posgresApp)
	if err != nil {
		return 500, err.Error()
	}

	err = client.Put().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Body(data).
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

func addSlaveAddress(addr resource.Address) (int, string) {
	client, _ := resource.GetApplicationClientScheme()
	var posgresApp resource.Application
	err := client.Get().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Do().Into(&posgresApp)
	if err != nil {
		return 500, fmt.Sprintf("Error getting app obj %s", err.Error())
	}
	newAppObj := resource.Application{}
	if len(posgresApp.Status.Addresses.Slaves) == 0 {
		posgresApp.Status.Addresses.Slaves = []resource.Address{}
	}

	alreadyExists := false

	for i := range posgresApp.Status.Addresses.Slaves {
		if posgresApp.Status.Addresses.Slaves[i].IP == addr.IP {
			if posgresApp.Status.Addresses.Slaves[i].Port == addr.Port {
				posgresApp.Status.Addresses.Slaves[i].LastUpdated = addr.LastUpdated
				alreadyExists = true
				break

			}
		}
	}

	if !alreadyExists {
		posgresApp.Status.Addresses.Slaves = append(posgresApp.Status.Addresses.Slaves, addr)
	}

	data, err := json.Marshal(posgresApp)
	if err != nil {
		return 500, err.Error()
	}

	err = client.Put().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Body(data).
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

func updateMasterAddress(addr resource.Address) (int, string) {
	client, _ := resource.GetApplicationClientScheme()
	var posgresApp resource.Application
	err := client.Get().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Do().Into(&posgresApp)
	if err != nil {
		return 500, fmt.Sprintf("Error getting app obj %s", err.Error())
	}
	newAppObj := resource.Application{}

	posgresApp.Status.Addresses.Master.IP = addr.IP
	posgresApp.Status.Addresses.Master.Port = addr.Port
	posgresApp.Status.Addresses.Master.LastUpdated = addr.LastUpdated

	data, err := json.Marshal(posgresApp)

	err = client.Put().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Body(data).
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

func ScaleApp(scale int) (int, string) {
	kClient := client.GetClient()
	slaveDep, err := kClient.ExtensionsV1beta1().Deployments(apiv1.NamespaceDefault).Get("slave", metav1.GetOptions{})
	if err != nil {
		return 500, fmt.Sprintf("Error getting slave deployment %s", err.Error())
	}
	if int32(scale) == *slaveDep.Spec.Replicas {
		return 200, fmt.Sprintf("Scale is already %d", scale)
	}
	toSet := int32(scale)
	slaveDep.Spec.Replicas = &toSet
	updatedDep, err := kClient.ExtensionsV1beta1().Deployments(apiv1.NamespaceDefault).Update(slaveDep)
	if err != nil {
		return 500, fmt.Sprintf("Error updating slave deployment %s", err.Error())
	}
	if *updatedDep.Spec.Replicas == toSet {
		updated, _ := json.Marshal(updatedDep)
		return 200, fmt.Sprintf(string(updated))
	}
	return 419, fmt.Sprintf("Could not update scale")
}

func ResetSlaves() (int, string) {
	client, _ := resource.GetApplicationClientScheme()
	var posgresApp resource.Application
	err := client.Get().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Do().Into(&posgresApp)
	if err != nil {
		return 500, fmt.Sprintf("Error getting app obj %s", err.Error())
	}
	newAppObj := resource.Application{}

	posgresApp.Status.Addresses.Slaves = []resource.Address{}

	data, err := json.Marshal(posgresApp)
	if err != nil {
		return 500, err.Error()
	}

	err = client.Put().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Body(data).
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

func UpdateState(state string) (int, string) {
	client, _ := resource.GetApplicationClientScheme()
	var posgresApp resource.Application
	err := client.Get().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Do().Into(&posgresApp)
	if err != nil {
		return 500, fmt.Sprintf("Error getting app obj %s", err.Error())
	}
	newAppObj := resource.Application{}

	posgresApp.Status.State = state

	data, err := json.Marshal(posgresApp)
	if err != nil {
		return 500, err.Error()
	}

	err = client.Put().
		Resource(resource.AppResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name("posgres").
		Body(data).
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
