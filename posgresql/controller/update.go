package controller

import (
	"github.com/caascade/posgreSQL/posgresql/resource"

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

}
