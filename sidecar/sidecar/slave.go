package sidecar

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/caascade/posgreSQL/sidecar/client"
	"github.com/caascade/posgreSQL/sidecar/cmdline"
)

func InitSlave(input *cmdline.CmdlineArgs) {
	masterIP, _ := client.GetMasterAddress(input.ControllerIP, input.ControllerPort)
	log.Errorf("obtained master IP")
	cmd := exec.Command("cp", "-r", "/resources/.", "/var/lib/postgresql/data/")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error copying resources to the right dir %v", err)
	}

	cmd = exec.Command("sed", "-i", fmt.Sprintf("s/SLAVE_IP_ADDR/%s/g", masterIP), "/var/lib/postgresql/data/pg_hba.conf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error updating pg_hba.conf %s %v %+v", out, err, cmd)
	}

	f, err := os.OpenFile("/var/lib/postgresql/data/recovery.conf", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("Error creating recovery conf %v", err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("%s\n", "standby_mode = 'on'"))
	f.WriteString(fmt.Sprintf("primary_conninfo = 'host=%s port=5432 user=rep password=sid'\n", masterIP))
	f.WriteString(fmt.Sprintf("trigger_file = '/tmp/postgresql.trigger.5432'\n"))

	cmd = exec.Command("chown", "-R", "999:999", "/var/lib/postgresql/data")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error updating user id to 999 %s %v", out, err)
	}

	cmd = exec.Command("chmod", "0700", "/var/lib/postgresql/data")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error chmoding to 0700 %s %v", out, err)
	}

}
