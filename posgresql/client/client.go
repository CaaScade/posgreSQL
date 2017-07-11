package client

import (
	log "github.com/Sirupsen/logrus"
	"github.com/caascade/posgreSQL/posgresql/executor"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type _ interface {
	GetClient() *kubernetes.Clientset
	GetConfig() *rest.Config
}

var client *kubernetes.Clientset
var registration_uuid string
var global_config *rest.Config

const task = "init-client"

func Init(uuid, kubeconf string, inCluster bool) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		initClient(kubeconf, inCluster)
		executor.ReturnToken(task, uuid)
	}()
}

func GetClient() *kubernetes.Clientset {
	return client
}

func GetConfig() *rest.Config {
	return global_config
}

func initClient(kubeConf string, inCluster bool) {
	errChan := make(chan error, 0)
	doneChan := make(chan bool, 0)

	getConfig := func() *rest.Config {
		if inCluster {
			log.Info("Obtaining in-cluster config")
			config, err := rest.InClusterConfig()
			if err != nil {
				log.Errorf("Error obtaining in-cluster config")
				errChan <- err
				return nil
			}
			return config
		}
		config, err := clientcmd.BuildConfigFromFlags("", kubeConf)
		if err != nil {
			log.Errorf("Error obtaining out-of-cluster config")
			errChan <- err
			return nil
		}
		global_config = config
		return config
	}

	go func() {
		log.Info("Obtaining client for kube cluster")
		var err error
		config := getConfig()
		if config == nil {
			return
		}
		client, err = kubernetes.NewForConfig(config)
		if err != nil {
			log.Errorf("Error obtaining client for kube cluster")
			errChan <- err
			return
		}
		doneChan <- true
	}()

	// Experiment: Separate error and data pipelines
	select {
	case err := <-errChan:
		executor.SetErrorState(registration_uuid, err)
	case <-doneChan:
	}

}
