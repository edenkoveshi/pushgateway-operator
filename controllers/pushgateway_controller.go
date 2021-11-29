/*
Copyright 2021.

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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/resources"
	"github.com/prometheus-operator/pushgateway-operator/internal/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// PushgatewayReconciler reconciles a Pushgateway object
type PushgatewayReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	DefaultImage string
}

//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=pushgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=pushgateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=pushgateways/finalizers,verbs=update
func (r *PushgatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &monitoringv1alpha1.Pushgateway{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Pushgateway")
		return ctrl.Result{}, err
	}

	res, err := r.ReconcilePushgateway(instance, ctx)

	if err != nil {
		return ctrl.Result{}, err
	}

	return res, nil
}

func (r *PushgatewayReconciler) ReconcilePushgateway(pgw *monitoringv1alpha1.Pushgateway, ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	prometheus, err := r.GetPrometheus(pgw, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if prometheus != nil {
		pgw.Status.Prometheus = fmt.Sprintf("%s/%s", prometheus.Namespace, prometheus.Name)
		if prometheus.Spec.ServiceMonitorSelector != nil {
			pgw.Status.PrometheusServiceMonitorSelector = prometheus.Spec.ServiceMonitorSelector
		}
		err := r.Status().Update(ctx, pgw)
		if err != nil {
			logger.Error(err, util.LogMessage(pgw, "Failed to update status"))
		}
		logger.Info(fmt.Sprintf("%s/%s set up with Prometheus %s", pgw.Namespace, pgw.Name, pgw.Status.Prometheus))
	} else {
		pgw.Status.Prometheus = "N/A"
		err := r.Status().Update(ctx, pgw)
		if err != nil {
			logger.Error(err, util.LogMessage(pgw, "Failed to update status"))
		}
		logger.Info(fmt.Sprintf("No Prometheus instance found for %s/%s", pgw.Namespace, pgw.Name))
	}

	pgw.Status.Image = resources.GetImageOrDefault(pgw, r.DefaultImage)
	r.Status().Update(ctx, pgw)

	res, err := r.reconcilePushgatewayDeployment(pgw, ctx)

	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info(util.LogMessage(pgw, "Successfully reconciled deployment"))

	nres, err := r.reconcilePushgatewayService(pgw, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info(util.LogMessage(pgw, "Successfully reconciled Service"))
	res = util.UpdateReconcileResult(res, nres)

	nres, err = r.reconcilePushgatewayServiceMonitor(pgw, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info(util.LogMessage(pgw, "Successfully reconciled ServiceMonitor"))
	res = util.UpdateReconcileResult(res, nres)

	return res, nil
}

// GetPrometheus returns a Prometheus instance for the Pushgateway.
// If Spec.Prometheus is set, the instance will be looked according to
// to it. Otherwise, default Prometheus is returned.
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses,verbs=get;list
func (r *PushgatewayReconciler) GetPrometheus(pgw *monitoringv1alpha1.Pushgateway, ctx context.Context) (*monitoringv1.Prometheus, error) {
	if pgw.Spec.Prometheus == nil {
		// Default Prometheus will not return an error
		// Worst case scenario is its' not found and no Prometheus instance is set
		// manager can still continue
		return r.GetDefaultPrometheus(pgw, ctx), nil
	}

	if pgw.Spec.Prometheus.Name == "" {
		errmsg := fmt.Sprintf("instance %s/%s,Prometheus name cannot be empty", pgw.Namespace, pgw.Name)
		return nil, errors.New(errmsg)
	}

	prometheus := &monitoringv1.Prometheus{}
	prometheusName := pgw.Spec.Prometheus.Name
	prometheusNamespace := pgw.Namespace
	if pgw.Spec.Prometheus.Namespace != "" {
		prometheusNamespace = pgw.Spec.Prometheus.Namespace
	}

	err := r.Get(ctx, types.NamespacedName{Name: prometheusName, Namespace: prometheusNamespace}, prometheus)

	if err != nil {
		return nil, err
	}

	return prometheus, nil
}

// GetDefaultPrometheus returns the default prometheus in the current namespace
// of the Pushgateway. If more than one exists, it fails.
func (r *PushgatewayReconciler) GetDefaultPrometheus(pgw *monitoringv1alpha1.Pushgateway, ctx context.Context) *monitoringv1.Prometheus {
	logger := log.FromContext(ctx)
	promList := &monitoringv1.PrometheusList{}
	listOpts := []client.ListOption{
		client.InNamespace(pgw.Namespace),
	}

	if err := r.List(ctx, promList, listOpts...); err != nil {
		logger.Error(err, "Failed to list Prometheus", "Pushgateway.Namespace", pgw.Namespace, "Pushgateway.Name", pgw.Name)
		return nil
	}

	if len(promList.Items) == 0 {
		logger.Error(nil, "configuration invalid: No Prometheuses found in namespace", "Pushgateway.Namespace", pgw.Namespace, "Pushgateway.Name", pgw.Name)
		return nil
	}

	if len(promList.Items) > 1 {
		logger.Error(nil, "configuration invalid: More than one Prometheus exists in namespace", "Pushgateway.Namespace", pgw.Namespace, "Pushgateway.Name", pgw.Name)
		return nil
	}

	return promList.Items[0]
}

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

// SetupWithManager sets up the controller with the Manager.
func (r *PushgatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Pushgateway{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Complete(r)
}
