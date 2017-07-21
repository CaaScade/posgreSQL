package resource

import (
	log "github.com/Sirupsen/logrus"
	"github.com/caascade/posgreSQL/posgresql/client"
	"github.com/caascade/posgreSQL/posgresql/executor"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type _ interface {
	GetApplicationClientScheme() (*rest.RESTClient, *runtime.Scheme)
}

const task = "create-resource"

var registration_uuid string

var SchemeGroupVersion = schema.GroupVersion{Group: AppResourceGroup, Version: "v1"}

var appClient *rest.RESTClient
var appScheme *runtime.Scheme

func Init(uuid string) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		log.Info("initializating task: create-resource")
		createResource()
		executor.ReturnToken(task, uuid)
	}()
}

func createResource() {
	kClient := client.GetClient()

	if kClient == nil {
		return
	}

	app := &v1beta1.ThirdPartyResource{
		ObjectMeta: metav1.ObjectMeta{
			Name: "application.appextensions.k8s.io",
		},
		Versions: []v1beta1.APIVersion{
			{
				Name: "v1",
			},
		},
		Description: "Application Resource",
	}
	_, err := kClient.ExtensionsV1beta1().ThirdPartyResources().Create(app)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Errorf("Error creating application resource")
		executor.SetErrorState(registration_uuid, err)
		return
	}
	initClientScheme()
	//add retries or a sleep here?
	_, err = appClient.Get().Namespace(apiv1.NamespaceAll).Resource(AppResourcePlural).DoRaw()
	if err != nil {
		executor.SetErrorState(registration_uuid, err)
		return
	}
}

func GetApplicationClientScheme() (*rest.RESTClient, *runtime.Scheme) {
	return appClient, appScheme
}

func initClientScheme() {
	addToScheme := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion,
			&Application{},
			&ApplicationList{},
		)
		metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
		return nil
	}
	schemeBuilder := runtime.NewSchemeBuilder(addToScheme)
	scheme := runtime.NewScheme()
	if err := schemeBuilder.AddToScheme(scheme); err != nil {
		executor.SetErrorState(registration_uuid, err)
	}
	cfg := client.GetConfig()
	if cfg == nil {
		log.Fatalf("Error getting client, it is nil")
	}
	cfg.GroupVersion = &SchemeGroupVersion
	cfg.APIPath = "/apis"
	cfg.ContentType = runtime.ContentTypeJSON
	cfg.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	client, err := rest.RESTClientFor(cfg)
	if err != nil {
		executor.SetErrorState(registration_uuid, err)
	}
	appClient = client
	appScheme = scheme
}
