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
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

// SetupWithManager sets up the controller with the Manager.
func (r *PushgatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Pushgateway{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		//Watches(&source.Kind{Type: &monitoringv1.Prometheuses{}}, handler.EnqueueRequestsFromMapFunc(r.watchPrometheuses)).
		//Watches(&source.Kind{Type: &batchv1.Job{}}, r.watchJobs).
		Complete(r)
}
