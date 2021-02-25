package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var storedClientset *kubernetes.Clientset
var storedConfigMap *v1.ConfigMap

var namespace string
var configmapName string

func getClientSet() *kubernetes.Clientset {
	if storedClientset != nil {
		return storedClientset
	}

	if home := homedir.HomeDir(); home != "" {
		kubeconfig := flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		storedClientset = clientset
		return clientset
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	storedClientset = clientset
	return clientset
}

// InitializeInformer does what it says on the tin
func InitializeInformer(cm string) {
	configmapName = cm
	clientset := getClientSet()

	ns, err := GetNamespace()
	if err != nil {
		panic(err.Error())
	}
	namespace = ns

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, time.Second*10)
	informer := factory.Core().V1().ConfigMaps().Informer()

	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: onAdd,
		// When a pod gets updated
		UpdateFunc: onUpdate,
		// When a pod gets deleted
		DeleteFunc: onDelete,
	})
	log.Printf("Starting informer goroutine")
	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	<-stopper
}

func storeConfigMap(configmap *v1.ConfigMap) {
	storedConfigMap = configmap
}

func onAdd(obj interface{}) {
	configmap := obj.(*v1.ConfigMap)

	if configmap.GetNamespace() == namespace && configmap.GetName() == configmapName {
		log.Printf("Configmap [%s] Namespace [%s] Updated", configmapName, namespace)
		storeConfigMap(configmap)
	}
}

func onUpdate(objOld interface{}, objNew interface{}) {
	configmap := objNew.(*v1.ConfigMap)

	if configmap.GetNamespace() == namespace && configmap.GetName() == configmapName {
		log.Printf("Configmap [%s] Namespace [%s] Updated", configmapName, namespace)
		storeConfigMap(configmap)
	}
}

func onDelete(obj interface{}) {
	configmap := obj.(*v1.ConfigMap)

	if configmap.GetNamespace() == namespace && configmap.GetName() == configmapName {
		log.Printf("Configmap [%s] Namespace [%s] Deleted", configmapName, namespace)
		storeConfigMap(nil)
	}
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

// GetConfigMap returns the currently stored configmap
func GetConfigMap() *v1.ConfigMap {
	return storedConfigMap
}
