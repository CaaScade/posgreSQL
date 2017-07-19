package cmdline

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

type CmdlineArgs struct {
	ModeInitMaster bool
	ModeInitSlave  bool
	ModeSidecar    bool

	ControllerIP   string
	ControllerPort int
}

func ScanCmdline() (*CmdlineArgs, error) {
	var args CmdlineArgs
	args.addFlags()
	args.scan()
	err := args.validate()
	return &args, err
}

func (args *CmdlineArgs) addFlags() {
	flag.BoolVar(&args.ModeInitMaster, "init-master", false, "initializes master node")
	flag.BoolVar(&args.ModeInitSlave, "init-slave", false, "initializes slave node")
	flag.BoolVar(&args.ModeSidecar, "sidecar", false, "runs as the postgres sidecar")
	flag.StringVar(&args.ControllerIP, "controller-address", "", "The address of the application controller")
	flag.IntVar(&args.ControllerPort, "controller-port", 8080, "The port on which the controller is listening")
}

func (args *CmdlineArgs) scan() {
	flag.Parse()
}

func (args *CmdlineArgs) validate() error {
	if args.ModeInitMaster && args.ModeInitSlave {
		return fmt.Errorf("Should only specify one option")
	}

	if args.ModeInitSlave && args.ModeSidecar {
		return fmt.Errorf("Should only specify one option")
	}

	if args.ModeSidecar && args.ModeInitMaster {
		return fmt.Errorf("Should only specify one option")
	}

	if !(args.ModeInitMaster || args.ModeInitSlave || args.ModeSidecar) {
		return fmt.Errorf("Should specify atleast one option (master|slave|sidecar)")
	}

	if controllerAddr := os.Getenv("CAASCADE_CONTROLLER_ADDRESS"); controllerAddr != "" {
		if args.ControllerIP != "" {
			args.ControllerIP = controllerAddr
		}
	}

	if !args.ModeSidecar && args.ControllerIP == "" {
		return fmt.Errorf("Caascade controller IP is not set")
	}

	return nil
}
