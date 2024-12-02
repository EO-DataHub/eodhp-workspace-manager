package k8s

import (
	"context"
	"fmt"
	"reflect"
	"time"

	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// InitializeManager initializes and returns a Kubernetes manager
func InitializeManager() (manager.Manager, error) {
	// Create a new runtime scheme
	scheme := runtime.NewScheme()

	// Register the Workspace CRD
	if err := workspacev1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to register Workspace CRD scheme: %w", err)
	}

	// Create the manager
	k8sMgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes manager: %w", err)
	}

	log.Info().Msg("Kubernetes manager initialized")
	return k8sMgr, nil
}

// ProcessWorkspace processes a WorkspaceSettings pulsar message payload
func ProcessWorkspace(ctx context.Context, client client.Client, c *utils.Config, payload models.WorkspaceSettings) error {
	switch payload.Status {
	case "creating":
		return CreateWorkspace(ctx, client, payload, c)
	case "updating":
		return UpdateWorkspace(ctx, client, payload, c)
	case "deleting":
		return DeleteWorkspace(ctx, client, payload)
	default:
		return fmt.Errorf("unknown status: %s", payload.Status)
	}
}

// ListenForWorkspaceStatusUpdates listens for updates to the Workspace CRD
func ListenForWorkspaceStatusUpdates(ctx context.Context, mgr manager.Manager, statusUpdates chan models.WorkspaceStatus) error {
	informer, err := mgr.GetCache().GetInformer(ctx, &workspacev1alpha1.Workspace{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create informer for Workspace CRD")
		return err
	}

	// Add event handler to the informer
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			handleUpdate(oldObj, newObj, statusUpdates)
		},
	})

	log.Info().Msg("Workspace CRD informer started")
	return nil
}

// handleUpdate handles updates to the Workspace CRD
func handleUpdate(oldObj, newObj interface{}, statusUpdates chan models.WorkspaceStatus) {

	oldWorkspace, ok := oldObj.(*workspacev1alpha1.Workspace)
	if !ok {
		log.Error().Msg("Failed to cast old object to Workspace")
		return
	}

	// Cast new object to Workspace
	newWorkspace, ok := newObj.(*workspacev1alpha1.Workspace)
	if !ok {
		log.Error().Msg("Failed to cast new object to Workspace")
		return
	}

	// Check if the status has actually changed
	if reflect.DeepEqual(oldWorkspace.Status, newWorkspace.Status) {
		// If status hasn't changed, ignore the event
		log.Debug().Msg("No changes in Workspace status; skipping")
		return
	}

	// Create a WorkspaceStatus object to send to the channel
	statusUpdate := models.WorkspaceStatus{
		Name:        newWorkspace.Name,
		Namespace:   newWorkspace.Status.Namespace,
		AWS:         newWorkspace.Status.AWS,
		LastUpdated: time.Now().UTC(),
	}

	// Send the status update to the channel
	select {
	case statusUpdates <- statusUpdate:
		log.Info().Msgf("Status update sent to channel: %v", statusUpdate)
	default:
		log.Warn().Msg("Status updates channel is full; dropping update")
	}

}
