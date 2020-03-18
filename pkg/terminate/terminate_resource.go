package terminate

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/xcoulon/kubectl-terminate/pkg/logger"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// Terminate terminates the resource with the given type and name, ie, it removes
// all pending finalizers and deletes it afterwards
func Terminate(kind, namespace, name string, kubeconfigReader io.Reader, log logger.Logger) error {
	kubeconfig, err := newKubeConfig(kubeconfigReader)
	if err != nil {
		return err
	}
	discoveryClient, err := newDiscoveryClient(kubeconfig)
	if err != nil {
		return err
	}
	apiresource, err := lookupAPIResource(kind, discoveryClient)
	if err != nil {
		return err
	}
	cl, err := newResourceClient(kubeconfig, namespace, apiresource)
	if err != nil {
		return err
	}
	resource, err := cl.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	err = removeFinalizers(resource)
	if err != nil {
		return err
	}
	log.Debug("updating resource '%s'", resource.GetName())
	resource, err = cl.Update(resource, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	log.Debug("deleting resource '%s'", resource.GetName())
	if err := cl.Delete(resource.GetName(), &metav1.DeleteOptions{}); !errors.IsNotFound(err) {
		// do not ignore errors unless it's a "NotFound" error, which may happen
		// because the resource was scheduled for deletion and the update to remove its finalizer
		// (see above) was enough to trigger its deletion
		return err
	}
	return nil
}

func newKubeConfig(r io.Reader) (clientcmd.ClientConfig, error) {
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return clientcmd.NewClientConfigFromBytes(d)
}

func newDiscoveryClient(kubeconfig clientcmd.ClientConfig) (*discovery.DiscoveryClient, error) {
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return discovery.NewDiscoveryClientForConfig(config)
}

// find the API for the given resource type
func lookupAPIResource(n string, cl discovery.DiscoveryInterface) (metav1.APIResource, error) {
	apiResourceLists, err := cl.ServerPreferredResources()
	if err != nil {
		return metav1.APIResource{}, err
	}
	for _, rl := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(rl.GroupVersion)
		if err != nil {
			return metav1.APIResource{}, err
		}
		for _, r := range rl.APIResources {
			if r.Name == n || // eg: 'checlusters'
				r.SingularName == n || // eg: 'checluster'
				r.Name+"."+gv.Group == n || // eg: 'checlusters.org.eclipse.che'
				r.SingularName+"."+gv.Group == n { // eg: 'checluster.org.eclipse.che'
				r.Group = gv.Group
				r.Version = gv.Version
				return r, nil
			}
			for _, sn := range r.ShortNames {
				if sn == n {
					r.Group = gv.Group
					r.Version = gv.Version
					return r, nil
				}
			}
		}
	}
	return metav1.APIResource{}, fmt.Errorf("unknown resource type: '%s'", n)
}

func newResourceClient(kubeconfig clientcmd.ClientConfig, namespace string, apiresource metav1.APIResource) (dynamic.ResourceInterface, error) {
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	i, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	r := i.Resource(schema.GroupVersionResource{
		Group:    apiresource.Group,
		Version:  apiresource.Version,
		Resource: apiresource.Name,
	})
	if !apiresource.Namespaced {
		return r, nil
	}
	if namespace != "" {
		return r.Namespace(namespace), nil
	}
	ns, _, err := kubeconfig.Namespace()
	if err != nil {
		return nil, err
	}
	return r.Namespace(ns), nil
}

// checkResource verifies that the given resource meets the expected criteria
func checkResource(r *unstructured.Unstructured) error {
	if r == nil {
		return fmt.Errorf("missing resource to check")
	}
	finalizers, found, err := unstructured.NestedStringSlice(r.UnstructuredContent(), "metadata", "finalizers")
	if err != nil {
		return err
	}
	if !found || len(finalizers) == 0 {
		return MissingFinalizerError{name: r.GetName()}
	}
	return nil
}

func removeFinalizers(r *unstructured.Unstructured) error {
	err := checkResource(r)
	if err != nil && IsMissingFinalizerError(err) {
		return nil // do not modify the existing resource
	} else if err != nil {
		return err
	}
	return unstructured.SetNestedSlice(r.Object, []interface{}{}, "metadata", "finalizers") // set an empty slice to override the current value
}

// MissingFinalizerError the error to return during the resource check when the latter has not 'kubernetes' finalizer
type MissingFinalizerError struct {
	name string
}

func (e MissingFinalizerError) Error() string {
	return fmt.Sprintf("resource '%s' has no finalizers in its metadata", e.name)
}

// IsMissingFinalizerError returns 'true' if the given error is a MissingFinalizerError
func IsMissingFinalizerError(err error) bool {
	_, is := err.(MissingFinalizerError)
	return is
}
