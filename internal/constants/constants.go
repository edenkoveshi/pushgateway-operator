package constants

// Naming conventions
const (
	ContainerName        = "pushgateway"
	DeploymentSuffix     = "-pushgateway"
	ServiceSuffix        = "-pushgateway"
	ServiceMonitorSuffix = "-pushgateway"
	PortName             = "web"
)

// Default values
const (
	DefaultPort          = 9091
	DefaultTelemetryPath = "/metrics"
	DefaultImage         = "prom/pushgateway"
)

// Image arguments
const (
	EnableAdminAPIArg  = "--web.enable-admin-api"
	EnableLifecycleArg = "--web.enable-lifecycle"
	ListenAddressArg   = "--web.listen-address="
	TelemetryPathArg   = "--web.telemetry-path="
	LogLevelArg        = "--log.level="
	LogFormatArg       = "--log.format="
)

// k8s resources names
const (
	ResourceDeployment     = "Deployment"
	ResourceService        = "Service"
	ResourceServiceMonitor = "ServiceMonitor"
)

const (
	SERVICE_MONITOR_MAX_REPEAT   = 100 //To avoid infinite looping
	JOB_WAIT_TIME_SECONDS        = 10  // Wait for 10 seconds before trying to create Job again
	JOB_CREATION_TIMEOUT_SECONDS = 6   // Timeout after 1 min
)

const (
	PushgatewayEnvVar    = "PUSHGATEWAY"
	PushgatewayLabelName = "inject-pushgateway"
)

func PushgatewayLabels() map[string]string {
	return map[string]string{
		"role": "pushgateway",
	}
}
