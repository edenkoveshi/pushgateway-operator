/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PushgatewaySpec defines the desired state of Pushgateway
type PushgatewaySpec struct {
	// Image to use for the Pushgateway. If omitted, default image defined in the operator
	// environment variable pushgateway-default-base-image
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=1
	// +optional
	Image string `json:"image,omitempty"`

	// Prometheus instance to bind to. If left empty, the operator assumes there's
	// a single Prometheus instance in the same namespace as the Pushgateway.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=2
	// +optional
	Prometheus *PushgatewayPrometheus `json:"prometheus,omitempty"`

	// If set to true, Pushgateway will run as a sidecar container in the Prometheus instance.
	// Otherwise runs as an independant pod. Default is false.
	// +kubebuilder:default=false
	// +optional
	InjectAsSidecar bool `json:"injectAsSidecar,omitempty"`

	// How many replicas of the Pushgateway to run. This applies only if InjectAsSidecar
	// is set to false. Default is 1.
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Whether or not to enable Pushgateway admin API
	// Default is false.
	// +kubebuilder:default=false
	// +optional
	EnableAdminAPI bool `json:"enableAdminAPI,omitempty"`

	// Port to listen on.
	// Default port is 9091.
	// +optional
	Port int32 `json:"port,omitempty"`

	// Path to push and expose metrics on.
	// Defaults to /metrics.
	// +optional
	TelemetryPath string `json:"telemetryPath,omitempty"`

	// Whether or not to enable Pushgateway lifecycle
	// Sets the --web.enable-lifecycle options
	// +optional
	EnableLifecycle bool `json:"enableLifecycle,omitempty"`

	// Sets the log level for the exporter.
	// Must be one of: debug,info,warn or error
	// Default is info
	// +kubebuilder:validation:Enum={debug,info,warn,error}
	// +optional
	LogLevel string `json:"logLevel,omitempty"`

	// Sets the log format for the exporter.
	// Must be either logfmt or json
	// Default is logfmt
	// +kubebuilder:validation:Enum={logfmt,json}
	// +optional
	LogFormat string `json:"logFormat,omitempty"`

	// Override or change some of the created Service Monitor properties
	// Properties that cannot be overriden: Name, Port, Path, Scheme, HonorLabels and HonorTimestamps
	// Those can be configured in the relevant fields
	// +optional
	ServiceMonitorOverrides *ServiceMonitorOverride `json:"serviceMonitorOverrides,omitempty"`

	/*
		TODO:
		Add override for: metadata, deployment name, service name
		Add persistence
	*/
}

// PushgatewayPrometheus is the Prometheus instance linked to the Pushgateway. If
// Metrics will be scraped by this Prometheus and, if chosen, Pushgateway will be
// injected as a sidecar to this Prometheus.
type PushgatewayPrometheus struct {
	// Prometheus instance name.
	Name string `json:"name"`

	// Prometheus instance namespace. If left empty, current namespace is used.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

type ServiceMonitorOverride struct {
	// Override the Service Monitor object metadata
	// New metadata will be added to auto-generated metadata
	// In case of a collision, override will take over
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=1
	// +optional
	ObjectMeta *metav1.ObjectMeta `json:"metadataOverrides,omitempty"`

	// ServiceMonitor Endpoint configuration
	// +optional
	Endpoint *monitoringv1.Endpoint `json:"endpointOverrides,omitempty"`
}

// PushgatewayStatus defines the observed state of Pushgateway
type PushgatewayStatus struct {
	Prometheus                       string                `json:"prometheus,omitempty"`
	PrometheusServiceMonitorSelector *metav1.LabelSelector `json:"prometheusServiceMonitorSelector,omitempty"`
	Image                            string                `json:"image,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Prometheus",type="string",JSONPath=".status.prometheus",description="Pushgateway's Prometheus instance"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// Pushgateway is the Schema for the pushgateways API
type Pushgateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PushgatewaySpec   `json:"spec,omitempty"`
	Status PushgatewayStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PushgatewayList contains a list of Pushgateway
type PushgatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pushgateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pushgateway{}, &PushgatewayList{})
}
