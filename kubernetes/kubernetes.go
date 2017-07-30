package kubernetes

import (
	devconf "github.com/ricjcosme/kube-monkey/config"
	kube "k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/clientcmd"
)


// InCluster provides the configuration for the k8s client
// InCluster or OutofCluster for dev purposes
// based on the InCluster directive in the toml config file
// Type: bool
// Default: true
func InCluster(InCluster bool) (*rest.Config, error) {
	if InCluster {
		cfg, e := rest.InClusterConfig()
		return cfg, e
	} else {
		cfg, e := clientcmd.BuildConfigFromFlags("", devconf.KubeConfigPath())
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
