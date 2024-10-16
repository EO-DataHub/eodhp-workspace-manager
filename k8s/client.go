package k8s

import (
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func CreateK8sClient() (client.Client, error) {

	// Create the controller-runtime client configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	// Register the Workspace CRD schema
	s := scheme.Scheme
	err = workspacev1alpha1.AddToScheme(s)
	if err != nil {
		return nil, err
	}

	// Create the controller-runtime client
	k8sClient, err := client.New(cfg, client.Options{Scheme: s})
	if err != nil {
		return nil, err
	}

	return k8sClient, nil

}
