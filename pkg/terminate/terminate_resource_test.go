package terminate

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xcoulon/kubectl-terminate/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
)

func TestLookupAPIResource(t *testing.T) {

	// given
	kubeconfig, server := setup(t)
	defer server.Close()
	client, err := newDiscoveryClient(kubeconfig)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {

		t.Run("core resource type", func(t *testing.T) {

			t.Run("by plural name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("namespaces", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Kind:       "Namespace",
					Name:       "namespaces",
					ShortNames: []string{"ns"},
					Namespaced: false,
					Version:    "v1",
				}, r)
			})

			t.Run("by short name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("ns", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Kind:       "Namespace",
					Name:       "namespaces",
					ShortNames: []string{"ns"},
					Namespaced: false,
					Version:    "v1",
				}, r)
			})
		})

		t.Run("custom resource type", func(t *testing.T) {

			t.Run("by unqualified singular name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("customtype", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Group:        "domain",
					Version:      "v1beta1",
					Name:         "customtypes",
					SingularName: "customtype",
					ShortNames:   []string{"ct"},
					Namespaced:   true,
					Kind:         "CustomType",
				}, r)
			})

			t.Run("by qualified singular name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("customtype.domain", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Group:        "domain",
					Version:      "v1beta1",
					Name:         "customtypes",
					SingularName: "customtype",
					ShortNames:   []string{"ct"},
					Namespaced:   true,
					Kind:         "CustomType",
				}, r)
			})

			t.Run("by plural name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("customtypes", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Group:        "domain",
					Version:      "v1beta1",
					Name:         "customtypes",
					SingularName: "customtype",
					ShortNames:   []string{"ct"},
					Namespaced:   true,
					Kind:         "CustomType",
				}, r)
			})

			t.Run("by unqualified plural name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("customtypes", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Group:        "domain",
					Version:      "v1beta1",
					Name:         "customtypes",
					SingularName: "customtype",
					ShortNames:   []string{"ct"},
					Namespaced:   true,
					Kind:         "CustomType",
				}, r)
			})

			t.Run("by qualified plural name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("customtypes.domain", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Group:        "domain",
					Version:      "v1beta1",
					Name:         "customtypes",
					SingularName: "customtype",
					ShortNames:   []string{"ct"},
					Namespaced:   true,
					Kind:         "CustomType",
				}, r)
			})

			t.Run("by short name", func(t *testing.T) {
				// when
				r, err := lookupAPIResource("ct", client)
				// then
				require.NoError(t, err)
				assert.Equal(t, metav1.APIResource{
					Group:        "domain",
					Version:      "v1beta1",
					Name:         "customtypes",
					SingularName: "customtype",
					ShortNames:   []string{"ct"},
					Namespaced:   true,
					Kind:         "CustomType",
				}, r)
			})
		})
	})

	t.Run("failures", func(t *testing.T) {

		t.Run("unknown resource type", func(t *testing.T) {
			// when
			_, err := lookupAPIResource("bar", client)
			// then
			require.Error(t, err)
			assert.Equal(t, err.Error(), "unknown resource type: 'bar'")
		})

	})
}

func TestFetchResource(t *testing.T) {

	// given
	kubeconfig, server := setup(t)
	defer server.Close()

	t.Run("ok", func(t *testing.T) {

		t.Run("namespace", func(t *testing.T) {
			// given
			cl, err := newResourceClient(kubeconfig, "bar", metav1.APIResource{
				Group:      "",
				Version:    "v1",
				Kind:       "Namespace",
				Name:       "namespaces",
				ShortNames: []string{"ns"},
			})
			require.NoError(t, err)
			// when
			actual, err := cl.Get("foo", metav1.GetOptions{})
			// then
			require.NoError(t, err)
			require.NotNil(t, actual)
			expected, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Spec: corev1.NamespaceSpec{
					Finalizers: []corev1.FinalizerName{
						corev1.FinalizerKubernetes,
					},
				},
				Status: corev1.NamespaceStatus{
					Phase: "Terminating",
				},
			})
			require.NoError(t, err)
			assert.Equal(t, expected, actual.Object)
		})
	})

	t.Run("failures", func(t *testing.T) {

		t.Run("unknown resource", func(t *testing.T) {
			// given
			cl, err := newResourceClient(kubeconfig, "bar", metav1.APIResource{
				Group:      "",
				Version:    "v1",
				Kind:       "Namespace",
				Name:       "namespaces",
				Namespaced: false,
				ShortNames: []string{"ns"},
			})
			require.NoError(t, err)
			// when
			_, err = cl.Get("unknown", metav1.GetOptions{})
			// then
			require.Error(t, err)
			require.IsType(t, &errors.StatusError{}, err)
			assert.True(t, errors.IsNotFound(err))

		})
	})
}
func TestRemoveFinalizers(t *testing.T) {

	t.Run("ok", func(t *testing.T) {

		t.Run("pod", func(t *testing.T) {
			// given
			object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "bar",
					Name:      "foo",
					Finalizers: []string{
						"custom",
					},
				},
				Spec: corev1.PodSpec{},
				Status: corev1.PodStatus{
					Phase: "Terminating",
				},
			})
			require.NoError(t, err)
			actual := &unstructured.Unstructured{
				Object: object,
			}
			// when
			err = removeFinalizers(actual)
			// then
			require.NoError(t, err)
			assert.Empty(t, actual.GetFinalizers())
		})
	})

	t.Run("failures", func(t *testing.T) {

		t.Run("missing finalizers", func(t *testing.T) {
			// given
			actual, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "bar",
					Name:       "foo",
					Finalizers: []string{},
				},
				Spec: corev1.PodSpec{},
				Status: corev1.PodStatus{
					Phase: "running",
				},
			})
			require.NoError(t, err)
			// when
			err = checkResource(&unstructured.Unstructured{
				Object: actual,
			})
			// then
			assert.IsType(t, err, MissingFinalizerError{})
		})
	})
}

func setup(t *testing.T) (clientcmd.ClientConfig, *httptest.Server) {
	server := test.NewServer(t)
	kubeconfigContent := test.NewKubeConfigContent(t, server.URL)
	kubeconfig, err := newKubeConfig(bytes.NewBuffer(kubeconfigContent))
	require.NoError(t, err)
	return kubeconfig, server
}
