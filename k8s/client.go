package k8s

import (
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// CreateK8sClients sets up both typed and dynamic Kubernetes clients
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

	// var config *rest.Config
	// var err error

	// // Check if we are running inside or outside the cluster
	// if _, exists := os.LookupEnv("KUBERNETES_SERVICE_HOST"); exists {
	// 	// In-cluster config
	// 	config, err = rest.InClusterConfig()
	// } else {
	// 	// Outside cluster, use kubeconfig
	// 	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	// 	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	// }
	// if err != nil {
	// 	return nil, nil, err
	// }

	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// dynamicClient, err := dynamic.NewForConfig(config)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// // Register Workspace CRD
	// err = workspacev1alpha1.AddToScheme(scheme.Scheme)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// return clientset, dynamicClient, nil
}
