package sidecar

import (
	"fmt"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/caascade/posgreSQL/sidecar/client"
	"github.com/caascade/posgreSQL/sidecar/cmdline"
)

func InitMaster(input *cmdline.CmdlineArgs) {
	client.UpdateMasterAddress(input.ControllerIP, input.ControllerPort)
	slaveIP, _ := client.GetSlaveAddress(input.ControllerIP, input.ControllerPort)
	log.Errorf("obtained slave IP")
	cmd := exec.Command("cp", "-r", "/resources/.", "/var/lib/postgresql/data/")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error copying resources to the right dir %v", err)
	}
	cmd = exec.Command("ls", "/var/lib/postgresql/data/")
	outx, _ := cmd.CombinedOutput()

	log.Errorf("%s", outx)

	cmd = exec.Command("sed", "-i", fmt.Sprintf("s/SLAVE_IP_ADDR/%s/g", slaveIP), "/var/lib/postgresql/data/pg_hba.conf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error updating pg_hba.conf %s %v %+v", out, err, cmd)
	}

	cmd = exec.Command("chown", "-R", "999:999", "/var/lib/postgresql/data")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error updating user id to 999 %v", err)
	}

	cmd = exec.Command("chmod", "0700", "/var/lib/postgresql/data")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error chmoding to 0700 %v", err)
	}

}
