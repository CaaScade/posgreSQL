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
	if input.RestoreBackup {
		getBackup()
	}

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

	cmd = exec.Command("rm", "-r", "/var/lib/postgresql/data/pg_log/postgresql.log")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error deleting old log file %v", err)
	}
}

func getBackup() {
	log.Infof("Getting backup from s3")
	cmd := exec.Command("/get-s3-backup.sh")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error downloading s3 file %s %v", out, err)
	}
	log.Infof("succesfully got the archive file %s", out)

	cmd = exec.Command("rm", "-r", "/resources")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error deleting old resources dir %s %v", out, err)
	}

	cmd = exec.Command("mkdir", "-p", "/resources")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error creating resources dir %s %v", out, err)
	}

	cmd = exec.Command("tar", "-xvf", "postgresql.tar", "-C", "/resources")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error un-tarring s3 file %s %v", out, err)
	}
	log.Infof("Succesfully copied backup %s", out)
}
