package k8s

import (
	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Client is a Kubernetes client
type Client struct {
	client client.Client
	Config *utils.Config
}

// NewClient creates a new Kubernetes client
func NewClient(appConfig *utils.Config) *Client {

	kubeConfig, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	// Add the Workspace CRD to the Kubernetes client
	s := scheme.Scheme
	err = workspacev1alpha1.AddToScheme(s)
	if err != nil {
		panic(err)
	}

	k8sClient, err := client.New(kubeConfig, client.Options{Scheme: s})
	if err != nil {
		panic(err)
	}

	return &Client{client: k8sClient, Config: appConfig}
}
