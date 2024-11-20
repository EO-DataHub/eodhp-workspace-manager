package k8s

import (
	"context"

	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// K8sInterface is an interface for the Kubernetes client
type K8sInterface interface {
	CreateWorkspace(ctx context.Context, req models.WorkspaceSettings) error
	UpdateWorkspace(ctx context.Context, req models.WorkspaceSettings) error
	DeleteWorkspace(ctx context.Context, payload models.WorkspaceSettings) error
}

// Client is a Kubernetes client
type K8sClient struct {
	client client.Client
	Config *utils.Config
}

// NewClient creates a new Kubernetes client
func NewClient(appConfig *utils.Config) *K8sClient {

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

	return &K8sClient{client: k8sClient, Config: appConfig}
}
