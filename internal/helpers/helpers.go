package helpers

import (
	"fmt"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	"k8s.io/apimachinery/pkg/labels"
)

func LogMessage(pgw *monitoringv1alpha1.Pushgateway, msg string) string {
	return fmt.Sprintf("%s/%s %s", pgw.Namespace, pgw.Name, msg)
}

func DeploymentName(name string) string {
	return fmt.Sprintf("%s%s", name, constants.DeploymentSuffix)
}

// Merge takes sets of labels and merges them. The last set
// provided will win in case of conflicts.
func Merge(sets ...map[string]string) labels.Set {
	merged := labels.Set{}
	for _, set := range sets {
		merged = labels.Merge(merged, set)
	}
	return merged
}
