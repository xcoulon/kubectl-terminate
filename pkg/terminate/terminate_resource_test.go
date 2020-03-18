package terminate

import (
	"bytes"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xcoulon/kubectl-terminate/pkg/logger"
	"github.com/xcoulon/kubectl-terminate/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTerminate(t *testing.T) {

	log := logger.NewLogger(os.Stdout, 1) // includes 'debug' messages

	t.Run("ok", func(t *testing.T) {

		t.Run("in default namespace", func(t *testing.T) {
			// given
			kubeconfig, server := setup(t)
			defer server.Close()
			// when
			err := Terminate("pod", "", "foo", kubeconfig, log)
			// then
			require.NoError(t, err)
		})

		t.Run("in other namespace", func(t *testing.T) {
			// given
			kubeconfig, server := setup(t)
			defer server.Close()
			// when
			err := Terminate("pod", "explicit", "foo", kubeconfig, log)
			// then
			require.NoError(t, err)
		})
	})
}

func TestLookupAPIResource(t *testing.T) {

	// given
	kubeconfigContent, server := setup(t)
	kubeconfig, err := newKubeConfig(kubeconfigContent)
	require.NoError(t, err)
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
	kubeconfigContent, server := setup(t)
	kubeconfig, err := newKubeConfig(kubeconfigContent)
	require.NoError(t, err)
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

func TestCheckResource(t *testing.T) {

	t.Run("pod with finalizer", func(t *testing.T) {
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
		err = checkResource(actual)
		// then
		require.NoError(t, err)
	})

	t.Run("pod without finalizer", func(t *testing.T) {
		// given
		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.Pod{
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
		actual := &unstructured.Unstructured{
			Object: object,
		}
		// when
		err = checkResource(actual)
		require.Error(t, err)
		assert.IsType(t, MissingFinalizerError{}, err)
		assert.Equal(t, "resource 'foo' has no finalizers in its metadata", err.Error())
	})
}

func TestRemoveFinalizers(t *testing.T) {

	t.Run("pod with finalizer", func(t *testing.T) {
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

	t.Run("pod without finalizer", func(t *testing.T) {
		// given
		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&corev1.Pod{
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
		actual := &unstructured.Unstructured{
			Object: object,
		}
		// when
		err = removeFinalizers(actual)
		require.NoError(t, err)
		assert.Empty(t, actual.GetFinalizers())
	})
}

func setup(t *testing.T) (io.Reader, *httptest.Server) {
	server := test.NewServer(t)
	kubeconfigContent := bytes.NewBuffer(test.NewKubeConfigContent(t, server.URL))
	return kubeconfigContent, server
}
