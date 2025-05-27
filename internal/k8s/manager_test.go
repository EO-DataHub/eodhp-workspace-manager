package k8s

import (
	"testing"
	"time"

	"github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleUpdate(t *testing.T) {
	ch := make(chan models.WorkspaceStatus, 1)

	oldObj := &v1alpha1.Workspace{
		Status: v1alpha1.WorkspaceStatus{
			State: "Pending",
		},
	}

	newObj := &v1alpha1.Workspace{
		ObjectMeta: oldObj.ObjectMeta,
		Status: v1alpha1.WorkspaceStatus{
			State:     "Running",
			Namespace: "ws-demo",
		},
	}

	handleUpdate(oldObj, newObj, ch)

	select {
	case msg := <-ch:
		assert.Equal(t, "Running", msg.State)
		assert.Equal(t, "ws-demo", msg.Namespace)
		assert.WithinDuration(t, time.Now().UTC(), msg.LastUpdated, time.Second)
	default:
		t.Fatal("expected status update but got none")
	}
}
