package resources

import (
	"fmt"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	"github.com/prometheus-operator/pushgateway-operator/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func SetOwnerReference(pgw *monitoringv1alpha1.Pushgateway) []metav1.OwnerReference {
	trueVar := true
	return []metav1.OwnerReference{
		metav1.OwnerReference{
			APIVersion: pgw.APIVersion,
			Kind:       pgw.Kind,
			Name:       pgw.Name,
			UID:        pgw.UID,
			Controller: &trueVar,
		},
	}
}

func PushgatewayLabels(pgw *monitoringv1alpha1.Pushgateway) labels.Set {
	return util.MergeLabels(constants.PushgatewayLabels(), pgw.Labels)
}

func GetImageOrDefault(pgw *monitoringv1alpha1.Pushgateway, defaultImage string) string {
	image := defaultImage
	if pgw.Spec.Image != "" {
		image = pgw.Spec.Image
	}
	return image
}

// Sets the port through spec.Port or default port
func GetPortOrDefault(pgw *monitoringv1alpha1.Pushgateway) int32 {
	port := int32(constants.DefaultPort)
	if pgw.Spec.Port > 0 {
		port = pgw.Spec.Port
	}
	return port
}

func GetTelemetryPathOrDefault(pgw *monitoringv1alpha1.Pushgateway) string {
	if pgw.Spec.TelemetryPath != "" {
		return fmt.Sprintf("%s%s", constants.TelemetryPathArg, pgw.Spec.TelemetryPath)
	}
	return constants.DefaultTelemetryPath
}
