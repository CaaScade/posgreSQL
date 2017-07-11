package app

import (
	log "github.com/Sirupsen/logrus"

	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type _ interface {
	GetApp() resource.Application
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
			Scale: 1,
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
