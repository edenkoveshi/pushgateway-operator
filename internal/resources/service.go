package resources

import (
	"fmt"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ServiceName(name string) string {
	return fmt.Sprintf("%s%s", name, constants.ServiceSuffix)
}

func PushgatewayService(pgw *monitoringv1alpha1.Pushgateway) *corev1.Service {
	port := GetPortOrDefault(pgw)
	labels := PushgatewayLabels(pgw)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ServiceName(pgw.Name),
			Namespace:       pgw.Namespace,
			Labels:          labels,
			OwnerReferences: SetOwnerReference(pgw),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       constants.PortName,
					Port:       port,
					TargetPort: intstr.FromString(constants.PortName),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector:  labels,
			ClusterIP: "None",
		},
	}
	return svc
}
