package resources

import (
	"fmt"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	"github.com/prometheus-operator/pushgateway-operator/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceMonitorName(name string) string {
	return fmt.Sprintf("%s%s", name, constants.ServiceMonitorSuffix)
}

func PushgatewayServiceMonitor(pgw *monitoringv1alpha1.Pushgateway) *monitoringv1.ServiceMonitor {
	truevar := true
	labels := PushgatewayLabels(pgw)

	// Find linked Prometheus instance label selector and merge it to the labels
	if promSmSelector := pgw.Status.PrometheusServiceMonitorSelector; promSmSelector != nil && pgw.Status.Prometheus != "N/A" {
		// Add labels under MatchLabels
		labels = util.MergeLabels(labels, promSmSelector.MatchLabels)

		//Iterate over MatchExpressions and make sure the labels match
		labels = HandleMatchExpressions(labels, promSmSelector.MatchExpressions)
	}

	/*
		TODO: Support creating ServiceMonitor in different namespace according
		to Prometheus' ServiceMonitorNamespaceSelector

		TODO: Support PodMonitor
	*/
	metadata := metav1.ObjectMeta{
		Name:      ServiceMonitorName(pgw.Name),
		Namespace: pgw.Namespace,
		Labels:    labels,
	}

	endpoint := &monitoringv1.Endpoint{}

	if override := pgw.Spec.ServiceMonitorOverrides; override != nil {
		if override.ObjectMeta != nil {
			util.MergeMetadata(&metadata, *override.ObjectMeta)
		}

		if override.Endpoint != nil {
			endpoint = override.Endpoint
		}
	}

	endpoint.Port = constants.PortName
	endpoint.Scheme = "http"
	endpoint.Path = GetTelemetryPathOrDefault(pgw)
	endpoint.HonorLabels = true
	endpoint.HonorTimestamps = &truevar

	// Create a basic service monitor
	svcmon := &monitoringv1.ServiceMonitor{
		ObjectMeta: metadata,
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: PushgatewayLabels(pgw),
			},
			Endpoints: []monitoringv1.Endpoint{
				*endpoint,
			},
		},
	}

	return svcmon
}

func HandleMatchExpressions(labels map[string]string, matchExp []metav1.LabelSelectorRequirement) map[string]string {
	for _, exp := range matchExp {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			if _, exists := labels[exp.Key]; !exists {
				//According to documentation Values must contain at least one element
				// This is a sure match
				labels[exp.Key] = exp.Values[0]
			}
		case metav1.LabelSelectorOpNotIn:
			if _, exists := labels[exp.Key]; !exists {
				// In this case we want the key to exist, but not to be in the Values array
				currentTime := time.Now()
				labels[exp.Key] = currentTime.Format("2006-01-02-15-04-05-000000") //Meaningless default value
			}
			if _, exists := labels[exp.Key]; exists {
				repeat := true
				i := 0
				for repeat && i < constants.SERVICE_MONITOR_MAX_REPEAT {
					repeat = false
					for _, val := range exp.Values {
						if val == labels[exp.Key] {
							currentTime := time.Now()
							labels[exp.Key] = currentTime.Format("2006-01-02-15-04-05-000000")
							repeat = true
						}
					}
					i++
				}
			}
		case metav1.LabelSelectorOpExists:
			if _, exists := labels[exp.Key]; !exists {
				// We want the key to exist with some meaningless value
				labels[exp.Key] = "true"
			}
		case metav1.LabelSelectorOpDoesNotExist:
			if _, exists := labels[exp.Key]; exists {
				// In this case we don't want the key to exist
				delete(labels, exp.Key)
			}
		}
	}
	return labels
}
