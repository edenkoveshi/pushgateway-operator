package deployment

import (
	"fmt"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	"github.com/prometheus-operator/pushgateway-operator/internal/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Creates a deployment for the Pushgateway
func PushgatewayDeployment(pgw *monitoringv1alpha1.Pushgateway, defaultImage string) *appsv1.Deployment {
	replicas := int32(1)
	if pgw.Spec.Replicas > 0 {
		replicas = pgw.Spec.Replicas
	}

	container := PushgatewayContainer(pgw, defaultImage)

	labels := helpers.Merge(constants.PushgatewayLabels(), pgw.Labels)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helpers.DeploymentName(pgw.Name),
			Namespace: pgw.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{*container},
				},
			},
			Replicas: &replicas,
		},
	}

	return dep
}

// Creates a container for the Pushgateway
func PushgatewayContainer(pgw *monitoringv1alpha1.Pushgateway, defaultImage string) *corev1.Container {
	image := getImageOrDefault(pgw, defaultImage)
	port := getPortOrDefault(pgw)
	args := getPushgatewayArgs(pgw, port)
	container := &corev1.Container{
		Name:  constants.ContainerName,
		Image: image,
		Args:  args,
		Ports: []corev1.ContainerPort{
			{
				Name:          constants.PortName,
				ContainerPort: port,
			},
		},
	}
	return container
}

// Sets the image through Spec.Image or environment variable
func getImageOrDefault(pgw *monitoringv1alpha1.Pushgateway, defaultImage string) string {
	image := defaultImage
	if pgw.Spec.Image != "" {
		image = pgw.Spec.Image
	}
	return image
}

// Sets the port through spec.Port or default port
func getPortOrDefault(pgw *monitoringv1alpha1.Pushgateway) int32 {
	port := int32(constants.DefaultPort)
	if pgw.Spec.Port > 0 {
		port = pgw.Spec.Port
	}
	return port
}

func getPushgatewayArgs(pgw *monitoringv1alpha1.Pushgateway, port int32) []string {
	arg := fmt.Sprintf("%s:%d", constants.ListenAddressArg, port)
	args := []string{arg}

	if pgw.Spec.TelemetryPath != "" {
		arg = fmt.Sprintf("%s:%s", constants.TelemetryPathArg, pgw.Spec.TelemetryPath)
		args = append(args, arg)
	}

	if pgw.Spec.EnableAdminAPI == true {
		args = append(args, constants.EnableAdminAPIArg)
	}

	if pgw.Spec.EnableLifecycle == true {
		args = append(args, constants.EnableLifecycleArg)
	}

	if pgw.Spec.LogLevel != "" {
		arg = fmt.Sprintf("%s%s", constants.LogLevelArg, pgw.Spec.LogLevel)
		args = append(args, arg)
	}

	if pgw.Spec.LogFormat != "" {
		arg = fmt.Sprintf("%s%s", constants.LogFormatArg, pgw.Spec.LogFormat)
		args = append(args, arg)
	}

	return args
}
