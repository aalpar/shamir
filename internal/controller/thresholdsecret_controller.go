package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	thresholdv1alpha1 "github.com/aalpar/shamir/api/v1alpha1"
)

// ThresholdSecretReconciler reconciles a ThresholdSecret object.
type ThresholdSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=threshold.shamir.io,resources=thresholdsecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=threshold.shamir.io,resources=thresholdsecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=threshold.shamir.io,resources=thresholdsecrets/finalizers,verbs=update

// Reconcile handles ThresholdSecret state transitions.
func (r *ThresholdSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO: implement state machine
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ThresholdSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&thresholdv1alpha1.ThresholdSecret{}).
		Named("thresholdsecret").
		Complete(r)
}
