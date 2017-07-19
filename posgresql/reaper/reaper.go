package reaper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caascade/posgreSQL/posgresql/app"
	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"

	log "github.com/Sirupsen/logrus"
)

const task = "reaper"

var registration_uuid string

func Init(uuid string) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		go start_reaping()
		executor.ReturnToken(task, uuid)
	}()
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
		time.Sleep(1 * time.Second)
	}
}

func checkAlive(addr resource.Address) bool {
	resp, err := http.Get(fmt.Sprintf("http://%s:9898/", addr.IP))
	if err != nil {
		log.Errorf("Error getting url for addr %s: %s", addr.IP, err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		log.Errorf("Status code is non-200; slave IP: %s, resp_code=%d", addr.IP, resp.StatusCode)
		return false
	}
	return true
}
