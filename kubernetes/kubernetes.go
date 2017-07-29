package kubernetes

import (
	devconf "github.com/ricjcosme/kube-monkey/config"
	kube "k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/clientcmd"
)

func InCluster(InCluster bool) (*rest.Config, error) {
	if InCluster {
		cfg, e := rest.InClusterConfig()
		return cfg, e
	} else {
		cfg, e := clientcmd.BuildConfigFromFlags("", "/Users/rc/.kube/config")
		return cfg, e
	}
}

func NewInClusterClient() (*kube.Clientset, error) {
	config, err := InCluster(devconf.InCluster())

	if err != nil {
		return nil, err
	}
	clientset, err := kube.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func VerifyClient(client *kube.Clientset) bool {
	_, err := client.ServerVersion()
	return err == nil
}
