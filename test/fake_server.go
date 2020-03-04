package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewServer returns a new HTTP Server which supports:
// - calls to `/api`
// - calls to `/apis`
// - calls on some predefined resources
// - 404 responses otherwise
// see https://github.com/kubernetes/client-go/blob/master/discovery/discovery_client_test.go
func NewServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var response interface{}
		fmt.Printf("processing %s %s\n", req.Method, req.URL)
		switch req.Method {
		case "GET":
			switch req.URL.Path {
			case "/api/v1":
				response = &metav1.APIResourceList{
					GroupVersion: "v1",
					APIResources: []metav1.APIResource{
						{
							Name:       "namespaces",
							ShortNames: []string{"ns"},
							Namespaced: false,
							Kind:       "Namespace",
						},
						{
							Name:         "pods",
							SingularName: "pod",
							ShortNames:   []string{"po"},
							Namespaced:   true,
							Kind:         "Pod",
						},
					},
				}
			case "/api":
				response = &metav1.APIVersions{
					Versions: []string{
						"v1",
					},
				}
			case "/apis":
				response = &metav1.APIGroupList{
					Groups: []metav1.APIGroup{
						{
							Name: "domain",
							Versions: []metav1.GroupVersionForDiscovery{
								{GroupVersion: "domain/v1beta1", Version: "v1beta1"},
							},
						},
					},
				}
			case "/apis/domain/v1beta1":
				response = &metav1.APIResourceList{
					GroupVersion: "domain/v1beta1",
					APIResources: []metav1.APIResource{
						{
							Name:         "customtypes",
							SingularName: "customtype",
							ShortNames:   []string{"ct"},
							Namespaced:   true,
							Kind:         "CustomType"},
					},
				}

			case "/api/v1/namespaces/foo":
				response = corev1.Namespace{
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
				}
			case "/api/v1/namespaces/default/pods/foo":
				response = corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "foo",
						Finalizers: []string{
							"cheesecake",
						},
					},
					Spec: corev1.PodSpec{},
					Status: corev1.PodStatus{
						Phase: "Terminating",
					},
				}
			case "/api/v1/namespaces/explicit/pods/foo":
				response = corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "explicit",
						Name:      "foo",
						Finalizers: []string{
							"cheesecake",
						},
					},
					Spec: corev1.PodSpec{},
					Status: corev1.PodStatus{
						Phase: "Terminating",
					},
				}
			case "/api/v1/namespaces/explicit/pods/bar": // no finalizer on this one
				response = corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "explicit",
						Name:      "foo",
					},
					Spec: corev1.PodSpec{},
					Status: corev1.PodStatus{
						Phase: "Terminating",
					},
				}
			default:
				fmt.Printf("object not found: %s %s\n", req.Method, req.URL)
				w.WriteHeader(http.StatusNotFound)
				return
			}
		case "PUT":
			switch req.URL.Path {
			case "/api/v1/namespaces/foo",
				"/api/v1/namespaces/default/pods/foo",
				"/api/v1/namespaces/explicit/pods/foo":
				// here we want to verify that the resource in the incoming request has no finalizer in its metadata
				// otherwise we return a 400 Bad Request error (unless there's something more appropriate?)
				data, err := ioutil.ReadAll(req.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error())) // nolint: errcheck
					return
				}
				pod := metav1.ObjectMeta{}
				err = json.Unmarshal(data, &pod)
				if err != nil {
					fmt.Printf("error while unmarshaling incoming request body: %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error())) // nolint: errcheck
					return
				}
				if len(pod.GetFinalizers()) > 0 {
					fmt.Printf("unexpected finalizers: %v\n", pod.GetFinalizers())
					w.WriteHeader(http.StatusBadRequest)
					// let's just return the request body in the response
					response, _ := ioutil.ReadAll(req.Body)
					w.Write(bytes.NewBuffer(response).Bytes()) // nolint: errcheck
					return
				}
				w.WriteHeader(http.StatusOK)
				// let's just return the request body in the response
				w.Write(data) // nolint: errcheck
				return
			}
		case "DELETE":
			switch req.URL.Path {
			case "/api/v1/namespaces/foo",
				"/api/v1/namespaces/default/pods/foo",
				"/api/v1/namespaces/explicit/pods/foo":
				// just accept the request
				w.WriteHeader(http.StatusNoContent)
				return
			}
		default:
			fmt.Printf("unexpected request: %s %s\n", req.Method, req.URL)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		output, err := json.Marshal(response)
		if err != nil {
			t.Errorf("unexpected encoding error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(output) // nolint: errcheck
	}))
}
