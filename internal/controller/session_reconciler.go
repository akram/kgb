package controller

import (
	"context"
	"time"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AgentSessionReconciler is the placeholder for drift remediation loops.
type AgentSessionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kgb.io,resources=agentsessions,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kgb.io,resources=agentsessions/status,verbs=get;update;patch

// Reconcile implements ctrl.Reconciler.
func (r *AgentSessionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var s kgbv1alpha1.AgentSession
	if err := r.Get(ctx, req.NamespacedName, &s); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if s.Status.LastTransitionTime == nil {
		now := metav1.Now()
		s.Status.LastTransitionTime = &now
		if err := r.Status().Update(ctx, &s); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			return ctrl.Result{}, err
		}
		logger.Info("initialized agentsession status", "name", s.Name)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers the reconciler.
func (r *AgentSessionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kgbv1alpha1.AgentSession{}).
		Complete(r)
}
