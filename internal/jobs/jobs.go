package jobs

import (
	"fmt"
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	"github.com/prometheus-operator/pushgateway-operator/internal/resources"
)

func InjectedJob(job *batchv1.Job, pgw *monitoringv1alpha1.Pushgateway) (*batchv1.Job, bool) {
	ret := job.DeepCopy()
	patched := []corev1.Container{}

	if len(ret.Spec.Template.Spec.Containers) == 0 {
		return ret, false
	}

	patchEnv := corev1.EnvVar{
		Name:  constants.PushgatewayEnvVar,
		Value: getJobPushPath(ret, pgw),
	}

	for _, container := range ret.Spec.Template.Spec.Containers {
		patchFlag := true
		for _, env := range container.Env {
			if reflect.DeepEqual(env, patchEnv) {
				patchFlag = false //do not patch container
			}
		}
		if patchFlag {
			container.Env = append(container.Env, patchEnv)
			patched = append(patched, container)
		}
	}

	updated := len(patched) > 0
	if updated {
		ret.Spec.Template.Spec.Containers = patched
		// Clean auto-generated fields
		ret.Spec.Selector = nil
		delete(ret.Spec.Template.Labels, "controller-uid")
		ret.ResourceVersion = ""
	}
	return ret, updated
}

func InjectedCronJob(job *batchv1.CronJob, pgw *monitoringv1alpha1.Pushgateway) (*batchv1.CronJob, bool) {
	ret := job.DeepCopy()
	patched := []corev1.Container{}

	if len(ret.Spec.JobTemplate.Spec.Template.Spec.Containers) == 0 {
		return ret, false
	}

	patchEnv := corev1.EnvVar{
		Name:  constants.PushgatewayEnvVar,
		Value: getCronJobPushPath(ret, pgw),
	}

	for _, container := range ret.Spec.JobTemplate.Spec.Template.Spec.Containers {
		patchFlag := true
		for _, env := range container.Env {
			if reflect.DeepEqual(env, patchEnv) {
				patchFlag = false //do not patch container
			}
		}
		if patchFlag {
			container.Env = append(container.Env, patchEnv)
			patched = append(patched, container)
		}
	}

	updated := len(patched) > 0
	if updated {
		ret.Spec.JobTemplate.Spec.Template.Spec.Containers = patched
		// Clean auto-generated fields
		ret.Spec.JobTemplate.Spec.Selector = nil
		delete(ret.Spec.JobTemplate.Labels, "controller-uid")
		ret.ResourceVersion = ""
	}
	return ret, updated
}

func getJobPushPath(job *batchv1.Job, pgw *monitoringv1alpha1.Pushgateway) string {
	return fmt.Sprintf("http://%s:%d%s/job/%s", resources.ServiceName(pgw), resources.GetPortOrDefault(pgw), resources.GetTelemetryPathOrDefault(pgw), job.Name)
}

func getCronJobPushPath(job *batchv1.CronJob, pgw *monitoringv1alpha1.Pushgateway) string {
	return fmt.Sprintf("http://%s:%d%s/job/%s", resources.ServiceName(pgw), resources.GetPortOrDefault(pgw), resources.GetTelemetryPathOrDefault(pgw), job.Name)
}

func IsJobInjectable(job *batchv1.Job) bool {
	labels := job.Labels
	for k := range labels {
		if k == constants.PushgatewayLabelName {
			return true
		}
	}
	return false
}

func IsCronJobInjectable(job *batchv1.CronJob) bool {
	labels := job.Labels
	for k := range labels {
		if k == constants.PushgatewayLabelName {
			return true
		}
	}
	return false
}
