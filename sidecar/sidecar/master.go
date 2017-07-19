package sidecar

import (
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/caascade/posgreSQL/sidecar/client"
	"github.com/caascade/posgreSQL/sidecar/cmdline"
)

func InitMaster(input *cmdline.CmdlineArgs) {
	client.ResetSlaves(input.ControllerIP, input.ControllerPort)
	client.UpdateMasterAddress(input.ControllerIP, input.ControllerPort)
	cmd := exec.Command("cp", "-r", "/resources/.", "/var/lib/postgresql/data/")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error copying resources to the right dir %v", err)
	}

	cmd = exec.Command("chown", "-R", "999:999", "/var/lib/postgresql/data")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error updating user id to 999 %v %s", err, out)
	}

	cmd = exec.Command("chmod", "0700", "/var/lib/postgresql/data")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error chmoding to 0700 %v", err)
	}

}
