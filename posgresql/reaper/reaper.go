package reaper

import (
	"strings"
	"time"

	"github.com/caascade/posgreSQL/posgresql/app"
	"github.com/caascade/posgreSQL/posgresql/client"
	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/Sirupsen/logrus"
)

const task = "reaper"

var registration_uuid string

func Init(uuid string) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		go start_reaping()
		go detect_master_failure()
		go start_slave_labelling()
		executor.ReturnToken(task, uuid)
	}()
}

func start_slave_labelling() {
	for {
		time.Sleep(10 * time.Second)
		kClient := client.GetClient()

		opts := metav1.ListOptions{}

		rsList, err := kClient.ExtensionsV1beta1().ReplicaSets(apiv1.NamespaceDefault).List(opts)
		if err != nil {
			continue
		}
		rsName := ""
		for _, rs := range rsList.Items {
			for _, ref := range rs.ObjectMeta.OwnerReferences {
				if ref.Name == "slave" && ref.Kind == "Deployment" {
					rsName = rs.ObjectMeta.Name
				}
			}
		}

		if rsName == "" {
			log.Infof("slave replica not created yet")
			continue
		}

		podList, err := kClient.CoreV1().Pods(apiv1.NamespaceDefault).List(opts)
		if err != nil {
			log.Error(err)
			continue
		}

		for _, pod := range podList.Items {
			for _, ref := range pod.ObjectMeta.OwnerReferences {
				if ref.Kind == "ReplicaSet" && ref.Name == rsName {
					if _, ok := pod.ObjectMeta.Labels[strings.Replace(pod.Status.PodIP, ".", "-", -1)]; !ok {
						log.Infof("updating label for slave pod %s", pod.ObjectMeta.Name)
						go update_pod_label(pod)
					}
				}
			}
		}
	}
}

func update_pod_label(pod apiv1.Pod) {
	kClient := client.GetClient()

	pod.ObjectMeta.Labels[(strings.Replace(pod.Status.PodIP, ".", "-", -1))] = "true"

	_, err := kClient.CoreV1().Pods(apiv1.NamespaceDefault).Update(&pod)
	if err != nil {
		log.Errorf("Error updating pod label %s", err.Error())
		return
	}
}

func detect_master_failure() {
	time.Sleep(90 * time.Second)
	for {
		appl := app.GetApp()
		alive := checkAlive(appl.Status.Addresses.Master)
		if !alive {
			kClient := client.GetClient()
			dep, _ := kClient.ExtensionsV1beta1().Deployments(apiv1.NamespaceDefault).Get("master", metav1.GetOptions{})
			if dep.ObjectMeta.Name == "master" {
				if appl.Status.State == "Recovery" {
					continue
				}
				app.UpdateState("Recovery")
				log.Errorf("The king is dead!")
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func start_reaping() {
	for {
		appl := app.GetApp()
		for i := range appl.Status.Addresses.Slaves {
			alive := checkAlive(appl.Status.Addresses.Slaves[i])
			if !alive {
				log.Infof("Deleting inactive slave %s", appl.Status.Addresses.Slaves[i])
				app.DeleteSlaveAddress(appl.Status.Addresses.Slaves[i])
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func checkAlive(addr resource.Address) bool {
	elapsed := time.Now().Unix() - addr.LastUpdated
	if elapsed > 10 {
		return false
	}
	return true
}
