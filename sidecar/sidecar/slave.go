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

	cmd := exec.Command("chmod", "0700", "/var/lib/postgresql/data")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error chmoding to 0700 %v", err)
	}

	cmd = exec.Command("pg_basebackup", "-x", "-h", masterIP, "-U", "postgres", "-D", "/var/lib/postgresql/data/")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error getting base backup %s %v %+v", out, err, cmd)
	}

	f, err := os.OpenFile("/var/lib/postgresql/data/recovery.conf", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("Error creating recovery conf %v", err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("%s\n", "standby_mode = 'on'"))
	f.WriteString(fmt.Sprintf("primary_conninfo = 'host=%s port=5432 user=postgres'\n", masterIP))
	f.WriteString(fmt.Sprintf("trigger_file = '/var/lib/postgresql/data/postgresql.trigger.5432'\n"))

	cmd = exec.Command("rm", "-r", "/var/lib/postgresql/data/postgresql.trigger.5432")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error deleting trigger file %s: %s", err.Error(), out)
	}

	cmd = exec.Command("chown", "-R", "999:999", "/var/lib/postgresql/data")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error updating user id to 999 %v %s", err, out)
	}
	client.UpdateSlaveAddress(input.ControllerIP, input.ControllerPort)
}
