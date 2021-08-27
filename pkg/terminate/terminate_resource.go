package terminate

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/xcoulon/kubectl-terminate/pkg/logger"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// ResourceMetadata the metadata of the resource to delete
type ResourceMetadata struct {
	Kind      string
	Namespace string
	Name      string
}

// Terminate terminates the resource with the given type and name, ie, it removes
// all pending finalizers and deletes it afterwards
func Terminate(metadata []ResourceMetadata, kubeconfigReader io.Reader, log logger.Logger) error {
	kubeconfig, err := newKubeConfig(kubeconfigReader)
	if err != nil {
		return err
	}
	discoveryClient, err := newDiscoveryClient(kubeconfig)
	if err != nil {
		return err
	}
	for _, m := range metadata {
		log.Debug("loading API resource")
		apiresource, err := lookupAPIResource(m.Kind, discoveryClient, log)
		if err != nil {
			return err
		}
		log.Debug("initializing client")
		cl, err := newResourceClient(kubeconfig, m.Namespace, apiresource)
		if err != nil {
			return err
		}
		log.Debug("loading resource '%s/%s' in namespace '%s'", m.Kind, m.Name, m.Namespace)
		resource, err := cl.Get(m.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		log.Debug("removing finalizers on '%s/%s'", resource.GetKind(), resource.GetName())
		err = removeFinalizers(resource)
		if err != nil {
			return err
		}
		log.Debug("updating '%s/%s'", resource.GetKind(), resource.GetName())
		resource, err = cl.Update(resource, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		log.Debug("deleting '%s/%s'", resource.GetKind(), resource.GetName())
		if err := cl.Delete(resource.GetName(), &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			// do not ignore errors unless it's a "NotFound" error, which may happen
			// because the resource was scheduled for deletion and the update to remove its finalizer
			// (see above) was enough to trigger its deletion
			return err
		}
		log.Info("%s \"%s\" terminated", m.Kind, m.Name)
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

var apiResourceCache = make(map[string]metav1.APIResource)

// find the API for the given resource type
func lookupAPIResource(n string, cl discovery.DiscoveryInterface, log logger.Logger) (metav1.APIResource, error) {
	if r, exists := apiResourceCache[n]; exists {
		return r, nil
	}
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
			log.Debug("checking API resource %s", spew.Sdump(r))
			if r.Name == n || // eg: 'checlusters'
				strings.ToLower(r.SingularName) == n || // eg: 'checluster'
				strings.ToLower(r.Kind) == n || // eg: 'checluster'
				r.Name+"."+gv.Group == n || // eg: 'checlusters.org.eclipse.che'
				r.SingularName+"."+gv.Group == n { // eg: 'checluster.org.eclipse.che'
				r.Group = gv.Group
				r.Version = gv.Version
				apiResourceCache[n] = r // keep in cache if we have multiple resource of the same kind to terminate
				return r, nil
			}
			for _, sn := range r.ShortNames {
				if sn == n {
					r.Group = gv.Group
					r.Version = gv.Version
					apiResourceCache[n] = r // keep in cache if we have multiple resource of the same kind to terminate
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
