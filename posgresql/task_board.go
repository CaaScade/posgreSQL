package posgresql

import (
	"github.com/pborman/uuid"

	"github.com/caascade/posgreSQL/posgresql/app"
	"github.com/caascade/posgreSQL/posgresql/client"
	"github.com/caascade/posgreSQL/posgresql/controller"
	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"
	"github.com/caascade/posgreSQL/posgresql/server"
)

var (
	// Tasks get executed in this order
	steps = []string{
		"init-client",
		"create-resource",
		"run-controller",
		"create-posgres-app",
		"server",
	}
)

func Exec(kubeconf string, inCluster bool, listenAddr string, listenPort int) error {
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
	server.Init(seedMap["server"].Uuid, listenAddr, listenPort)

	return executor.Exec(seedList)
}
