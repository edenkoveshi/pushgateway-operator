package resources

import (
	"fmt"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DeploymentName(name string) string {
	return fmt.Sprintf("%s%s", name, constants.DeploymentSuffix)
}

// Creates a deployment for the Pushgateway
// TODO: add requests/limits
func PushgatewayDeployment(pgw *monitoringv1alpha1.Pushgateway) *appsv1.Deployment {
	replicas := int32(1)
	if pgw.Spec.Replicas > 0 {
		replicas = pgw.Spec.Replicas
	}

	container := PushgatewayContainer(pgw)

	labels := PushgatewayLabels(pgw)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            DeploymentName(pgw.Name),
			Namespace:       pgw.Namespace,
			Labels:          labels,
			OwnerReferences: SetOwnerReference(pgw),
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
func PushgatewayContainer(pgw *monitoringv1alpha1.Pushgateway) *corev1.Container {
	image := pgw.Status.Image
	port := GetPortOrDefault(pgw)
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

func getPushgatewayArgs(pgw *monitoringv1alpha1.Pushgateway, port int32) []string {
	arg := fmt.Sprintf("%s:%d", constants.ListenAddressArg, port)
	args := []string{arg}

	if pgw.Spec.TelemetryPath != "" {
		arg = GetTelemetryPathOrDefault(pgw)
		args = append(args, arg)
	}

	if pgw.Spec.EnableAdminAPI {
		args = append(args, constants.EnableAdminAPIArg)
	}

	if pgw.Spec.EnableLifecycle {
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
