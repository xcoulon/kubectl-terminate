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
		Use:           "terminate (TYPE NAME | TYPE/NAME)",
		Short:         "removes the finalizers and deletes the given resource",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1), // can terminate mulitiple resources at once
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger(cmd.OutOrStdout(), loglevel)
			// look-up the kubeconfig to use
			kubeconfigFile, err := getKubeconfigFile(kubeconfig)
			if err != nil {
				return fmt.Errorf("error while locating KUBECONFIG: %w", err)
			}
			log.Debug("using kubeconfig at %s", kubeconfigFile.Name())
			// deal with resource kinds/names
			resources := make([]terminate.ResourceMetadata, 0, len(args))
			// if the first arg does not contain a `/`, then assume its a kind.
			// otherwise, split all args
			if !strings.Contains(args[0], "/") {
				kind := args[0]
				// all other args are the resource names (of the same kind)
				for _, name := range args[1:] {
					resources = append(resources, terminate.ResourceMetadata{
						Kind:      kind,
						Name:      name,
						Namespace: namespace,
					})
				}
			} else {
				for _, arg := range args {
					kindname := strings.Split(arg, "/")
					if len(kindname) != 2 {
						return fmt.Errorf("invalid resource name: %s", arg)
					}
					kind := kindname[0]
					name := kindname[1]
					resources = append(resources, terminate.ResourceMetadata{
						Kind:      kind,
						Name:      name,
						Namespace: namespace,
					})
				}
			}
			if err := terminate.Terminate(resources, kubeconfigFile, log); err != nil {
				return errors.Cause(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "", "", "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "(optional) the namespace scope for this CLI request")
	cmd.Flags().IntVarP(&loglevel, "loglevel", "v", 0, "log level for V logs (set to 1 or higher to display DEBUG messages)")

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
