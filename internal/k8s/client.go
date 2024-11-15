package k8s

import (
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	client client.Client
}

func NewClient() *Client {

	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	s := scheme.Scheme
	err = workspacev1alpha1.AddToScheme(s)
	if err != nil {
		panic(err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: s})
	if err != nil {
		panic(err)
	}

	return &Client{client: k8sClient}
	// // Create the controller-runtime client configuration
	// cfg, err := config.GetConfig()
	// if err != nil {
	// 	return nil, err
	// }

	// // Register the Workspace CRD schema
	// s := scheme.Scheme
	// err = workspacev1alpha1.AddToScheme(s)
	// if err != nil {
	// 	return nil, err
	// }

	// // Create the controller-runtime client
	// k8sClient, err := client.New(cfg, client.Options{Scheme: s})
	// if err != nil {
	// 	return nil, err
	// }

	// return k8sClient, nil

}
