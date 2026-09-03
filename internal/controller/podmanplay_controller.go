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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	kastellanuseurminddev1 "github.com/kastellan/kastellan/api/v1"
	"github.com/kastellan/kastellan/pkg/agentprotocol/server"
)

// PodmanPlayReconciler reconciles a PodmanPlay object
type PodmanPlayReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	statusProcessor *server.StatusProcessor
}

// SetStatusProcessor sets the status processor for workload status updates.
func (r *PodmanPlayReconciler) SetStatusProcessor(sp *server.StatusProcessor) {
	r.statusProcessor = sp
}

// +kubebuilder:rbac:groups=kastellan.useurmind.de,resources=podmanplays,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kastellan.useurmind.de,resources=podmanplays/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kastellan.useurmind.de,resources=podmanplays/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodmanPlay object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *PodmanPlayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var play kastellanuseurminddev1.PodmanPlay
	if err := r.Get(ctx, req.NamespacedName, &play); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Update status from workload status if available
	if r.statusProcessor != nil {
		status, exists := r.statusProcessor.GetWorkloadStatus(play.Namespace, play.Name)
		if exists && status.Phase != "" {
			log.Info("Updating PodmanPlay status from workload status",
				"namespace", play.Namespace, "name", play.Name, "phase", status.Phase)

			now := metav1.Now()

			// Update observed generation
			play.Status.ObservedGeneration = play.Generation

			// Add host status (for now, just use the workload name as host)
			// In production, this would be mapped from workload to actual host
			if play.Status.Hosts == nil {
				play.Status.Hosts = []kastellanuseurminddev1.PodmanPlayHostStatus{}
			}

			hostStatus := kastellanuseurminddev1.PodmanPlayHostStatus{
				Name:               play.Name,
				Phase:              status.Phase,
				AppliedGeneration:  status.Generation,
				LastTransitionTime: &now,
			}

			// Check if host already exists and update or add
			found := false
			for i, h := range play.Status.Hosts {
				if h.Name == play.Name {
					play.Status.Hosts[i] = hostStatus
					found = true
					break
				}
			}

			if !found {
				play.Status.Hosts = append(play.Status.Hosts, hostStatus)
			}

			// Update conditions based on phase
			conditions := []metav1.Condition{}

			availableCondition := metav1.Condition{
				Type:               "Available",
				Status:             metav1.ConditionTrue,
				Reason:             "WorkloadReady",
				Message:            "Workload is ready",
				LastTransitionTime: &now,
			}

			if status.Phase != server.PhaseReady {
				availableCondition.Status = metav1.ConditionFalse
				availableCondition.Reason = "WorkloadNotReady"
				availableCondition.Message = "Workload is not ready"
			}

			conditions = append(conditions, availableCondition)
			play.Status.Conditions = conditions

			if err := r.Status().Update(ctx, &play); err != nil {
				log.Error(err, "Failed to update PodmanPlay status", "namespace", play.Namespace, "name", play.Name)
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodmanPlayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kastellanuseurminddev1.PodmanPlay{}).
		Named("podmanplay").
		Complete(r)
}
