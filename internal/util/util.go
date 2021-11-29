package util

import (
	"fmt"

	monitoringv1alpha1 "github.com/prometheus-operator/pushgateway-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
)

func LogMessage(pgw *monitoringv1alpha1.Pushgateway, msg string) string {
	return fmt.Sprintf("%s/%s %s", pgw.Namespace, pgw.Name, msg)
}

// Merge takes sets of labels and merges them. The last set
// provided will win in case of conflicts.
func MergeLabels(sets ...map[string]string) labels.Set {
	merged := labels.Set{}
	for _, set := range sets {
		merged = labels.Merge(merged, set)
	}
	return merged
}

// MergeMetadata takes labels and annotations from the old resource and merges
// them into the new resource. If a key is present in both resources, the new
// resource wins. It also copies the ResourceVersion from the old resource to
// the new resource to prevent update conflicts.
func MergeMetadata(new *metav1.ObjectMeta, old metav1.ObjectMeta) {
	if old.ResourceVersion != "" {
		new.ResourceVersion = old.ResourceVersion
	}

	new.SetLabels(MergeLabels(new.Labels, old.Labels))
	new.SetAnnotations(MergeLabels(new.Annotations, old.Annotations))
}

// UpdateReconcileResult creates a new Result based on the new and existing results provided to it.
// This includes setting "Requeue" to true in the Result if set to true in the new Result but not
// in the existing Result, while also updating RequeueAfter if the RequeueAfter value for the new
// result is less the the RequeueAfter value for the existing Result.
func UpdateReconcileResult(currResult, newResult ctrl.Result) ctrl.Result {

	if newResult.Requeue {
		currResult.Requeue = true
	}

	if newResult.RequeueAfter != 0 {
		if currResult.RequeueAfter == 0 || newResult.RequeueAfter < currResult.RequeueAfter {
			currResult.RequeueAfter = newResult.RequeueAfter
		}
	}

	return currResult
}
