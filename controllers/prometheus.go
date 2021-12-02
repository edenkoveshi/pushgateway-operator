package controllers

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
)

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
