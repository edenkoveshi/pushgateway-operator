package controllers

import (
	"context"
	"fmt"
	"time"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	"github.com/prometheus-operator/pushgateway-operator/internal/constants"
	"github.com/prometheus-operator/pushgateway-operator/internal/jobs"
	batchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CronJobReconciler reconciles a Job object
type CronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &batchv1.CronJob{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	res, err := r.ReconcileCronJob(instance, ctx)

	if err != nil {
		return ctrl.Result{}, err
	}

	return res, nil
}

// Reconcile CronJobs to inject them.
// Desired behaviour:
// If the label exists and environment variable does not, create it
// If the label does not exist don't do anything (if env was already defined before
// it will not be undefined, simply not reconciled)
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;update;list;patch;watch;delete;create;
func (r *CronJobReconciler) ReconcileCronJob(job *batchv1.CronJob, ctx context.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	pgw, err := r.GetPushgatewayInNamespace(job.Namespace, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if jobs.IsCronJobInjectable(job) {
		newJob, updateNeeded := jobs.InjectedCronJob(job, pgw)
		if updateNeeded {
			err := r.Delete(ctx, job)
			if err != nil {
				logger.Error(err, fmt.Sprintf("Failed to inject CronJob %s/%s", job.Namespace, job.Name))
				return ctrl.Result{}, err
			}

			err = r.Create(ctx, newJob)
			if err != nil {
				// It is very likely to get an error message that the Job
				// already exists because it hasn't been deleted yet.
				// Try again after {JOB_WAIT_TIME_SECONDS} seconds for a maximum of
				// {JOB_CREATION_TIMEOUT_SECONDS} seconds
				i := 0
				for k8serrors.IsAlreadyExists(err) && i < constants.JOB_CREATION_TIMEOUT_SECONDS {
					time.Sleep(constants.JOB_WAIT_TIME_SECONDS * time.Second)
					i++
					err = r.Create(ctx, newJob)
				}
				if err != nil { // Still getting an error.
					logger.Error(err, fmt.Sprintf("Failed to inject CronJob %s/%s", newJob.Namespace, newJob.Name))
					return ctrl.Result{}, err
				}
			}
			logger.Info(fmt.Sprintf("CronJob %s/%s successfully injected", newJob.Namespace, newJob.Name))
		}
	}

	return ctrl.Result{}, nil
}

func (r *CronJobReconciler) GetPushgatewayInNamespace(namespace string, ctx context.Context) (*monitoringv1alpha1.Pushgateway, error) {
	logger := log.FromContext(ctx)
	pgwList := &monitoringv1alpha1.PushgatewayList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
	}

	if err := r.List(ctx, pgwList, listOpts...); err != nil {
		logger.Error(err, "Failed to list Pushgateways", "Namespace", namespace)
		return nil, err
	}

	if len(pgwList.Items) == 0 {
		err := fmt.Errorf("no Pushgateways found in namespace %s", namespace)
		return nil, err
	}

	if len(pgwList.Items) > 1 {
		err := fmt.Errorf("more than 1 Pushgateway found in namespace %s", namespace)
		return nil, err
	}

	return &pgwList.Items[0], nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.CronJob{}).
		Complete(r)
}
