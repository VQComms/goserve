package main

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getClientSet() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// GetNamespace returns the current namespace
// Checks for ENV variable of POD_NAMESPACE from the downward api, if that doesnt exist
// It pulls the namespace from the serviceaccount
// If both of those fail, it returns an empty string
func GetNamespace() (string, error) {
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns, nil
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
		return "", err
	}

	return "", nil
}

// FetchConfigMap fetches a configmap in the current namespace
func FetchConfigMap(name string) (*v1.ConfigMap, error) {
	namespace, err := GetNamespace()
	if err != nil {
		return nil, err
	}
	clientset := getClientSet()

	return clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}
