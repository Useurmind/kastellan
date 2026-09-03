/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	kastellanuseurminddev1 "github.com/kastellan/kastellan/api/v1"
	"github.com/kastellan/kastellan/pkg/agentprotocol/server"
)

const (
	heartbeatTimeout = 2 * time.Minute
)

// ExternalHostReconciler reconciles a ExternalHost object
type ExternalHostReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	heartbeatServer *server.HeartbeatProcessor
}

// SetHeartbeatServer sets the heartbeat server for status updates.
func (r *ExternalHostReconciler) SetHeartbeatServer(hb *server.HeartbeatProcessor) {
	r.heartbeatServer = hb
}

// +kubebuilder:rbac:groups=kastellan.useurmind.de,resources=externalhosts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kastellan.useurmind.de,resources=externalhosts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kastellan.useurmind.de,resources=externalhosts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ExternalHost object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *ExternalHostReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var host kastellanuseurminddev1.ExternalHost
	if err := r.Get(ctx, req.NamespacedName, &host); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Update status from heartbeat data if available
	if r.heartbeatServer != nil {
		sessionID := req.Name
		if state, exists := r.heartbeatServer.GetSessionState(sessionID); exists {
			log.Info("Updating ExternalHost status from heartbeat", "host", req.Name)

			now := metav1.Now()
			connected := time.Since(state.LastHeartbeat) <= heartbeatTimeout

			// Update conditions
			conditions := []metav1.Condition{}

			connectedCondition := metav1.Condition{
				Type:               "Connected",
				Status:             metav1.ConditionTrue,
				Reason:             "AgentHeartbeatReceived",
				Message:            "Agent sent heartbeat recently",
				LastTransitionTime: now,
			}

			if !connected {
				connectedCondition.Status = metav1.ConditionFalse
				connectedCondition.Reason = "AgentHeartbeatTimeout"
				connectedCondition.Message = "No heartbeat received within timeout"
			}

			conditions = append(conditions, connectedCondition)

			readyCondition := metav1.Condition{
				Type:               "Ready",
				Status:             metav1.ConditionTrue,
				Reason:             "PodmanAvailable",
				Message:            "Podman runtime is available",
				LastTransitionTime: now,
			}

			if !state.RuntimeAvailable {
				readyCondition.Status = metav1.ConditionFalse
				readyCondition.Reason = "PodmanUnavailable"
				readyCondition.Message = "Podman runtime is unavailable"
			}

			conditions = append(conditions, readyCondition)

			// Update status
			host.Status.Conditions = conditions
			if err := r.Status().Update(ctx, &host); err != nil {
				log.Error(err, "Failed to update ExternalHost status", "host", req.Name)
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExternalHostReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kastellanuseurminddev1.ExternalHost{}).
		Named("externalhost").
		Complete(r)
}
