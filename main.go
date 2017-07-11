package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/caascade/posgreSQL/cmdline"
	"github.com/caascade/posgreSQL/constants"
	"github.com/caascade/posgreSQL/posgresql"
)

func main() {
	log.Infof("Starting %s Application", constants.APP_NAME)

	input, err := cmdline.ScanCmdline()
	if err != nil {
		log.Fatalf("Error initializing controller: %v", err)
	}

	if err := posgresql.Exec(input.Kubeconf, input.InCluster, input.ListenAddress, input.ListenPort); err != nil {
		log.Fatalf("Error running %s Application: %v", constants.APP_NAME, err)
	}
}
