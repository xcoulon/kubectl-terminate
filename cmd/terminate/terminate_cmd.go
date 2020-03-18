package terminate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xcoulon/kubectl-terminate/pkg/logger"
	"github.com/xcoulon/kubectl-terminate/pkg/terminate"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitAndExecute() {
	if err := NewCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}

func NewCommand() *cobra.Command {

	var kubeconfig string
	var namespace string
	var loglevel int

	cmd := &cobra.Command{
		Use:           "terminate",
		Short:         "removes the finalizers and deletes the given resource",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.RangeArgs(1, 2), // for now, accept in the form of `kind name` (not `kind/name`)
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger(cmd.OutOrStdout(), loglevel)
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
			log.Debug("using kubeconfig at %s", kubeconfigFile.Name())
			if err := terminate.Terminate(kind, namespace, name, kubeconfigFile, log); err != nil {
				return errors.Cause(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s \"%s\" terminated", kind, name)
			log.Info("")

			return nil
		},
	}
	cmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "", "", "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "(optional) the namespace scope for this CLI request")
	cmd.Flags().IntVarP(&loglevel, "loglevel", "v", 0, "log level for V logs")

	return cmd
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
