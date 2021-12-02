package controllers

import (
	"context"
	"reflect"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/resources"
	"github.com/prometheus-operator/pushgateway-operator/internal/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

/*
	TODO: Make a generic reconcile function (unstructured?)
*/

// Reconcile the deployment needed for the pushgateway
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;update;create;list;patch;watch;delete
func (r *PushgatewayReconciler) reconcilePushgatewayDeployment(pgw *monitoringv1alpha1.Pushgateway, ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	found := &appsv1.Deployment{}
	desired := resources.PushgatewayDeployment(pgw)

	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: pgw.Namespace}, found)

	//Deployment does not exist. Create it.
	if err != nil && k8serrors.IsNotFound(err) {
		return ctrl.Result{Requeue: true}, r.Create(ctx, desired)
	} else if err != nil {
		logger.Error(err, util.LogMessage(pgw, "Failed to get Deployment"))
		return ctrl.Result{}, err
	}

	// Check whether or not the deployment has been changed
	// If it has changed, reconcile it
	if !reflect.DeepEqual(desired.Spec, found.Spec) {
		util.MergeMetadata(&desired.ObjectMeta, found.ObjectMeta)
		return ctrl.Result{Requeue: true}, r.Update(ctx, desired)
	}

	return ctrl.Result{}, nil
}

// Reconcile the service needed for the pushgateway
// +kubebuilder:rbac:groups=*,resources=services,verbs=get;update;create;list;patch;watch;delete
func (r *PushgatewayReconciler) reconcilePushgatewayService(pgw *monitoringv1alpha1.Pushgateway, ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	found := &corev1.Service{}
	desired := resources.PushgatewayService(pgw)

	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: pgw.Namespace}, found)

	//Service does not exist. Create it.
	if err != nil && k8serrors.IsNotFound(err) {
		return ctrl.Result{Requeue: true}, r.Create(ctx, desired)
	} else if err != nil {
		logger.Error(err, util.LogMessage(pgw, "Failed to get Service"))
		return ctrl.Result{}, err
	}

	// Check whether or not the service has been changed
	// If it has changed, reconcile it
	if !reflect.DeepEqual(desired.Spec, found.Spec) {
		util.MergeMetadata(&desired.ObjectMeta, found.ObjectMeta)
		return ctrl.Result{Requeue: true}, r.Update(ctx, desired)
	}

	return ctrl.Result{}, nil
}

// Reconcile the service monitor needed for the pushgateway
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;update;create;list;patch;watch;delete
func (r *PushgatewayReconciler) reconcilePushgatewayServiceMonitor(pgw *monitoringv1alpha1.Pushgateway, ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	found := &monitoringv1.ServiceMonitor{}
	desired := resources.PushgatewayServiceMonitor(pgw)

	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: pgw.Namespace}, found)

	//ServiceMonitor does not exist. Create it.
	if err != nil && k8serrors.IsNotFound(err) {
		return ctrl.Result{Requeue: true}, r.Create(ctx, desired)
	} else if err != nil {
		logger.Error(err, util.LogMessage(pgw, "Failed to get ServiceMonitor"))
		return ctrl.Result{}, err
	}

	// Check whether or not the service has been changed
	// If it has changed, reconcile it
	if !reflect.DeepEqual(desired.Spec, found.Spec) {
		util.MergeMetadata(&desired.ObjectMeta, found.ObjectMeta)
		return ctrl.Result{Requeue: true}, r.Update(ctx, desired)
	}

	return ctrl.Result{}, nil
}
