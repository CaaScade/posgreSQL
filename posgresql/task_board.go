package posgresql

import (
	"github.com/pborman/uuid"

	"github.com/caascade/posgreSQL/posgresql/app"
	"github.com/caascade/posgreSQL/posgresql/block"
	"github.com/caascade/posgreSQL/posgresql/client"
	"github.com/caascade/posgreSQL/posgresql/controller"
	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"
)

var (
	// Tasks get executed in this order
	steps = []string{
		"init-client",
		"create-resource",
		"run-controller",
		"create-posgres-app",
		"block",
	}
)

func Exec(kubeconf string, inCluster bool) error {
	seedMap := map[string]executor.Token{}

	for i := range steps {
		// 1 uuid per serial task
		// parallel tasks share the same uuid
		seedMap[steps[i]] = executor.Token{
			Name:   steps[i],
			Uuid:   uuid.New(),
			Actors: map[string]bool{},
		}
	}

	seedList := []executor.Token{}

	for i := range steps {
		seedList = append(seedList, seedMap[steps[i]])
	}

	client.Init(seedMap["init-client"].Uuid, kubeconf, inCluster)
	resource.Init(seedMap["create-resource"].Uuid)
	controller.Init(seedMap["run-controller"].Uuid)
	app.Init(seedMap["create-posgres-app"].Uuid)
	block.Init(seedMap["block"].Uuid)

	return executor.Exec(seedList)
}
