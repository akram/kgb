package controller

import (
	"context"
	"time"

	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	"github.com/akram/kgb/pkg/approval"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ApprovalReconciler drives Kubernetes Events + optional webhooks for pending approvals.
type ApprovalReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	Notifier  approval.Notifier
}

// +kubebuilder:rbac:groups=kgb.io,resources=approvalrequests,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kgb.io,resources=approvalrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile implements ctrl.Reconciler.
func (r *ApprovalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var ar kgbv1alpha1.ApprovalRequest
	if err := r.Get(ctx, req.NamespacedName, &ar); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if ar.Status.Phase == "" {
		ar.Status.Phase = kgbv1alpha1.ApprovalPending
		if err := r.Status().Update(ctx, &ar); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if ar.Status.Phase == kgbv1alpha1.ApprovalPending && ar.Status.Message == "" {
		r.Recorder.Eventf(&ar, corev1.EventTypeWarning, "ApprovalPending",
			"Intent %s target=%s requires approval", ar.Spec.Intent.Action, ar.Spec.Intent.Target)
		if r.Notifier != nil {
			if err := r.Notifier.NotifyPending(ctx, &ar); err != nil {
				logger.Error(err, "approval notifier failed")
			}
		}
		ar.Status.Message = "notified"
		if err := r.Status().Update(ctx, &ar); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers the reconciler.
func (r *ApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kgbv1alpha1.ApprovalRequest{}).
		Complete(r)
}
