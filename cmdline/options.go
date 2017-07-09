package cmdline

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

type CmdlineArgs struct {
	Kubeconf  string
	InCluster bool
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
