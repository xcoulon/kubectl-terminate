package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xcoulon/kubectl-terminate/pkg/terminate"
)

// TerminateCmd the main command to remove all finalizers and delete a given resource
var TerminateCmd *cobra.Command

func init() {
	var kubeconfig string
	var namespace string
	TerminateCmd = &cobra.Command{
		Use:   "terminate",
		Short: "removes the finalizers and deletes the given resource",
		Args:  cobra.RangeArgs(1, 2), // for now, accept in the form of `kind name` (not `kind/name`)
		RunE: func(cmd *cobra.Command, args []string) error {
			var kind, name string
			if len(args) == 1 {
				args = strings.Split(args[0], "/")
			}
			kind = args[0]
			name = args[1]
			kubeconfigFile, err := getKubeconfigFile(kubeconfig)
			if err != nil {
				return fmt.Errorf("error while locating KUBECONFIG: %w", err)
			}
			err = terminate.Terminate(kind, namespace, name, kubeconfigFile)
			if err != nil {
				return fmt.Errorf("error while terminating resource: %w", err)
			}
			return nil
		},
	}
	TerminateCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "", "", "(optional) absolute path to the kubeconfig file")
	TerminateCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "(optional) the namespace scope for this CLI request")
}

// getKubeconfigFile returns a file reader on (by order of match):
// - the --kubeconfig CLI argument if it was provided
// - the $KUBECONFIG file it the env var was set
// - the <user_home_dir>/.kube/config file
func getKubeconfigFile(kubeconfig string) (*os.File, error) {
	var path string
	if kubeconfig != "" {
		path = kubeconfig
	} else if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		path = kubeconfig
	} else {
		path = filepath.Join(homeDir(), ".kube", "config")
	}
	return os.Open(path)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
