package block

import (
	log "github.com/Sirupsen/logrus"

	"github.com/caascade/posgreSQL/posgresql/executor"
)

const task = "block"

var registration_uuid string

func Init(uuid string) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		log.Info("Controller is running!")
	}()
}
