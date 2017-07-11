package cmdline

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

type CmdlineArgs struct {
	Kubeconf  string
	InCluster bool

	ListenAddress string
	ListenPort    int
}

func ScanCmdline() (*CmdlineArgs, error) {
	var args CmdlineArgs
	args.addFlags()
	args.scan()
	err := args.validate()
	return &args, err
}

func (args *CmdlineArgs) addFlags() {
	flag.StringVar(&args.Kubeconf, "kube-config", "", "absolute path to the kubeconf file for the cluster")
	flag.BoolVar(&args.InCluster, "in-cluster", false, "Specifies if the controller is running as a pod within the cluster")
	flag.StringVar(&args.ListenAddress, "listen-address", "0.0.0.0", "Specifies the address on which the server listens")
	flag.IntVar(&args.ListenPort, "listen-port", 8080, "Specifies the port on which the server listens")
}

func (args *CmdlineArgs) scan() {
	flag.Parse()
}

func (args *CmdlineArgs) validate() error {
	if args.InCluster {
		if args.Kubeconf != "" {
			return fmt.Errorf("in-cluster flag is set, kube-apiserver flag is unneccessary")
		}
		return nil
	}
	if args.Kubeconf == "" {
		return fmt.Errorf("out-of-cluster kube config file path is not specified")
	}
	return nil
}
