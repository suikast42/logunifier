package process

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/internal/streams/ingress"
	"github.com/suikast42/logunifier/pkg/model"
)

func ValidateAndFix(ecs *model.EcsLogEntry, msg *nats.Msg) {
	if !ecs.IsIngressSet() {
		ecs.AppendValidationError("Ingress is empty")
		ecs.SetIngress("Empty")
	}

	if !ecs.IsOrgNameSet() {
		ecs.AppendValidationError("No organisation name set")
		ecs.SetOrgName("NoOrg")
	}

	if !ecs.IsServiceNameSet() {
		ecs.AppendValidationError("Service name is empty")
		ecs.SetSetServiceName("Empty")
	}
	if !ecs.IsServiceTypeSet() {
		ecs.AppendValidationError("Service type is empty")
		ecs.SetServiceType("Empty")
	}

	if !ecs.IsLogLevelSet() {
		ecs.AppendValidationError("Log level not found")
		ecs.SetLogLevel(model.LogLevel_not_set)
	}

	if ecs.Timestamp == nil {
		ecs.AppendValidationError("Timestamp not found. Set to fallback")
		ecs.SetTimeStamp(ingress.TimestampFromIngestion(msg))
	}

	if !ecs.IsPatternSet() {
		ecs.AppendValidationError("No pattern found")
		ecs.SetPattern("NoPattern")
	}

	if !ecs.IsEnvironmentSet() {
		ecs.AppendValidationError("No environment set")
		ecs.SetEnvironment("NoEnv")
	}

	if !ecs.IsStackSet() {
		ecs.AppendValidationError("No stack set")
		ecs.SetStack("NoStack")
	}

	if !ecs.IsServiceNameSpaceSet() {
		ecs.AppendValidationError("No namespace set")
		ecs.SetServiceNameSpace("NoNameSpace")
	}

	if !ecs.IsHostNameSet() {
		ecs.AppendValidationError("No host name set")
		ecs.SetHostName("NoHost")
	}
	// Delete the debug info if there is no error occured there
	if !ecs.HasProcessError() {
		ecs.ProcessError = nil
	}

	if !ecs.HasValidationError() {
		ecs.ValidationError = nil
	}

	ecs.SetMarkerEmojis()
}
