package controller

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const task = "run-controller"

var registration_uuid string
var appClient *rest.RESTClient
var appScheme *runtime.Scheme

func Init(uuid string) {
	registration_uuid = uuid
	go func() {
		log.Info("waiting to run task: run-controller")
		executor.ObtainToken(task, uuid)
		go startController()
		executor.ReturnToken(task, uuid)
	}()
}

func startController() {
	appClient, appScheme = resource.GetApplicationClientScheme()

	source := cache.NewListWatchFromClient(
		appClient,
		resource.AppResourcePlural,
		apiv1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		source,
		&resource.Application{},
		30*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    addApp,
			UpdateFunc: updateApp,
			DeleteFunc: deleteApp,
		},
	)
	stopChan := make(chan struct{}, 0)
	controller.Run(stopChan)
}

func addApp(obj interface{}) {
	app := obj.(*resource.Application)
	appObj, err := appScheme.Copy(app)
	if err != nil {
		log.Errorf("Error copying created app %s: %v", app, err)
		return
	}

	log.Infof("Created app object %s", appObj.(*resource.Application).ObjectMeta.Name)
}

func updateApp(oldObj, newObj interface{}) {
	log.Infof("Updating app object %s", newObj.(*resource.Application).ObjectMeta.Name)
}

func deleteApp(obj interface{}) {
	log.Infof("Delete app %s", obj.(*resource.Application).ObjectMeta.Name)
}
