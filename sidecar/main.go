package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/caascade/posgreSQL/sidecar/cmdline"
	"github.com/caascade/posgreSQL/sidecar/sidecar"
)

func main() {
	log.Infof("starting sidecar")
	input, err := cmdline.ScanCmdline()
	if err != nil {
		log.Fatalf("error starting sidecar %v", err)
	}
	if input.ModeInitMaster {
		sidecar.InitMaster(input)
	}

	if input.ModeInitSlave {
		sidecar.InitSlave(input)
	}

	if input.ModeSidecar {
		sidecar.InitSidecar(input)
	}
}
