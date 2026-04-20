package controller

import (
	"context"
	"time"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PolicyReconciler validates Policy specs and mirrors generation into status.
type PolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kgb.io,resources=policies,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kgb.io,resources=policies/status,verbs=get;update;patch

// Reconcile implements ctrl.Reconciler.
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var pol kgbv1alpha1.Policy
	if err := r.Get(ctx, req.NamespacedName, &pol); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	msg := validatePolicy(&pol)
	if pol.Status.ObservedGeneration == pol.Generation && pol.Status.Message == msg {
		return ctrl.Result{}, nil
	}
	pol.Status.ObservedGeneration = pol.Generation
	pol.Status.Message = msg
	if err := r.Status().Update(ctx, &pol); err != nil {
		if apierrors.IsConflict(err) {
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	logger.Info("reconciled policy", "name", pol.Name, "message", msg)
	return ctrl.Result{}, nil
}

func validatePolicy(p *kgbv1alpha1.Policy) string {
	if p.Spec.Decision == kgbv1alpha1.PolicyTimeBoundAllow {
		if p.Spec.TimeBoundAllow == "" {
			return "invalid: time_bound_allow requires spec.timeBoundAllow duration"
		}
		if _, err := time.ParseDuration(p.Spec.TimeBoundAllow); err != nil {
			return "invalid: spec.timeBoundAllow must be a Go duration (e.g. 15m)"
		}
	}
	return "ok"
}

// SetupWithManager registers the reconciler.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kgbv1alpha1.Policy{}).
		Complete(r)
}
