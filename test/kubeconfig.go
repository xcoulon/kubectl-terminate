package test

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	kubeconfigTmpl = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: "{{ . }}"
  name: test-server
contexts:
- context:
    cluster: test-server
  name: test-server
current-context: test-server`
)

// NewKubeConfigFile returns the path to a the kubeconfig file to access
// the server with the given URL
func NewKubeConfigFile(t *testing.T, serverURL string) (home string, kubeconfig *os.File) {
	homeDir := os.TempDir()
	dotKubeDir := filepath.Join(homeDir, ".kube")
	err := os.MkdirAll(dotKubeDir, os.ModePerm)
	require.NoError(t, err)
	f, err := os.Create(filepath.Join(dotKubeDir, "config"))
	require.NoError(t, err)
	content := NewKubeConfigContent(t, serverURL)
	// fmt.Printf("kubeconfig: %s\n%s\n", f.Name(), string(content))
	_, err = f.Write(content)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	return homeDir, f
}

// NewKubeConfigContent returns an `io.Reader` to the kubeconfig to
// access the server with the given URL
func NewKubeConfigContent(t *testing.T, serverURL string) []byte {
	tmpl, err := template.New("kubeconfig").Parse(string(kubeconfigTmpl))
	require.NoError(t, err)
	r := bytes.NewBuffer(nil)
	err = tmpl.Execute(r, serverURL)
	require.NoError(t, err)
	return r.Bytes()
}
