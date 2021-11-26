package constants

// Naming conventions
const (
	ContainerName    = "pushgateway"
	DeploymentSuffix = "-pushgateway"
	ServiceSuffix    = "-pushgateway"
	PodMonitorSuffix = "-pushgateway"
	PortName         = "metrics"
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

func PushgatewayLabels() map[string]string {
	return map[string]string{
		"role": "pushgateway",
	}
}
