package terminate_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/xcoulon/kubectl-terminate/cmd/terminate"
	"github.com/xcoulon/kubectl-terminate/test"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminateCmd(t *testing.T) {

	// given
	server := test.NewServer(t)
	defer server.Close()

	oldKubeConfig := os.Getenv("KUBECONFIG")
	defer func() {
		if oldKubeConfig != "" {
			os.Setenv("KUBECONFIG", oldKubeConfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}()
	os.Unsetenv("KUBECONFIG")

	t.Run("ok", func(t *testing.T) {

		t.Run("with custom kubeconfig", func(t *testing.T) {

			t.Run("pod in current namespace", func(t *testing.T) {
				// given
				_, kubeconfig := test.NewKubeConfigFile(t, server.URL)
				defer os.Remove(kubeconfig.Name())
				// when
				out, err := executeCommand(terminate.NewCommand(), "--kubeconfig="+kubeconfig.Name(), "pod", "cookie")
				// then
				require.NoError(t, err)
				assert.Equal(t, "pod \"cookie\" terminated\n", out)

			})

			t.Run("pod in dessert namespace", func(t *testing.T) {
				// given
				_, kubeconfig := test.NewKubeConfigFile(t, server.URL)
				defer os.Remove(kubeconfig.Name())
				// when
				_, err := executeCommand(terminate.NewCommand(), "--kubeconfig="+kubeconfig.Name(), "--namespace=dessert", "pod", "cookie")
				// then
				require.NoError(t, err)
			})
		})

		t.Run("with envvar kubeconfig", func(t *testing.T) {

			t.Run("custom resource with splitted name", func(t *testing.T) {
				// given
				_, kubeconfig := test.NewKubeConfigFile(t, server.URL)
				oldKubeConfig := os.Getenv("KUBECONFIG")
				defer func() {
					if oldKubeConfig != "" {
						os.Setenv("KUBECONFIG", oldKubeConfig)
					} else {
						os.Unsetenv("KUBECONFIG")
					}
				}()
				os.Setenv("KUBECONFIG", kubeconfig.Name())
				// when
				_, err := executeCommand(terminate.NewCommand(), "pod", "cookie")
				// then
				require.NoError(t, err)
			})

		})

		t.Run("with userhome kubeconfig", func(t *testing.T) {

			// given
			homeDir, _ := test.NewKubeConfigFile(t, server.URL)
			oldHome := os.Getenv("HOME")
			defer func() {
				if oldHome != "" {
					os.Setenv("HOME", oldHome)
				} else {
					os.Unsetenv("HOME")
				}
			}()
			os.Setenv("HOME", homeDir)

			t.Run("single resource", func(t *testing.T) {

				t.Run("with compact name", func(t *testing.T) {
					// when
					_, err := executeCommand(terminate.NewCommand(), "pod/cookie")
					// then
					require.NoError(t, err)
				})

				t.Run("with splitted name", func(t *testing.T) {
					// when
					_, err := executeCommand(terminate.NewCommand(), "pod", "cookie")
					// then
					require.NoError(t, err)
				})
			})

			t.Run("multiple resources", func(t *testing.T) {

				t.Run("with compact names", func(t *testing.T) {
					// when
					_, err := executeCommand(terminate.NewCommand(), "pod/cookie", "deploy/latte")
					// then
					require.NoError(t, err)
				})

				t.Run("with splitted names", func(t *testing.T) {
					// when
					_, err := executeCommand(terminate.NewCommand(), "pod", "cookie", "cookie2")
					// then
					require.NoError(t, err)
				})
			})
		})

	})

	t.Run("failures", func(t *testing.T) {

		t.Run("with invalid kubeconfig", func(t *testing.T) {
			// given
			oldKubeConfig := os.Getenv("KUBECONFIG")
			defer func() {
				if oldKubeConfig != "" {
					os.Setenv("KUBECONFIG", oldKubeConfig)
				} else {
					os.Unsetenv("KUBECONFIG")
				}
			}()
			os.Setenv("KUBECONFIG", "invalid")
			// when
			_, err := executeCommand(terminate.NewCommand(), "pod", "cookie")
			// then
			require.Error(t, err)
			assert.Equal(t, "error while locating KUBECONFIG: open invalid: no such file or directory", err.Error())
		})
	})

}

// see https://github.com/spf13/cobra/blob/master/command_test.go#L16-L29
// nolint: unparam
func executeCommand(cmd *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	cmd.SetArgs(args)
	_, err = cmd.ExecuteC()
	return buf.String(), err
}
