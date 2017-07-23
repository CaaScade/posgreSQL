package sidecar

import (
	"os"
	"time"

	"github.com/caascade/posgreSQL/sidecar/client"
	"github.com/caascade/posgreSQL/sidecar/cmdline"
	"github.com/caascade/posgreSQL/sidecar/tail"

	log "github.com/Sirupsen/logrus"
)

func InitSidecar(input *cmdline.CmdlineArgs) {
	log.Infof("Starting sidecar for posgres %s", input.SidecarType)
	tail.InitTail("/var/lib/postgresql/data/pg_log/postgresql.log")
	go recoveryCheck(input.ControllerIP, input.ControllerPort)
	if input.SidecarType == "master" {
		go streamLogs(input.ControllerIP, input.ControllerPort)
	}

	for {
		switch input.SidecarType {
		case "master":
			client.UpdateMasterAddressNoPanic(input.ControllerIP, input.ControllerPort)
		case "slave":
			client.UpdateSlaveAddressNoPanic(input.ControllerIP, input.ControllerPort)
		}
		time.Sleep(5 * time.Second)
	}
}

func streamLogs(ip string, port int) {
	tailChan := tail.Tail()
	client.StreamLogs(ip, port, tailChan)
}

func recoveryCheck(ip string, port int) {
	for {
		state := client.GetState(ip, port)
		if state == "Recovery" {
			promote(ip, port)
		}
		time.Sleep(5 * time.Second)
	}
}

func promote(ip string, port int) {
	f, err := os.Create("/var/lib/postgresql/data/postgresql.trigger.5432")
	if err != nil {
		if err != os.ErrExist {
			log.Errorf("error creating trigger file")
			return
		}
	}
	f.Close()
	for {
		state := client.GetState(ip, port)
		if state == "Configured" {
			os.Remove("/var/lib/postgresql/data/postgresql.trigger.5432")
			os.Exit(1)
		}
		time.Sleep(5 * time.Second)
	}
}
