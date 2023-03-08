package process

import (
	"github.com/suikast42/logunifier/pkg/model"
)

func ValidateAndFix(ecs *model.EcsLogEntry, metalog *model.MetaLog) {
	if !ecs.IsJobNameSet() {
		ecs.AppendParseError("Job name is empty ")
		ecs.SetJobName("Empty")
	}
	if !ecs.IsJobTypeSet() {
		ecs.AppendParseError("Job type is empty")
		ecs.SetJobType("Empty")
	}
	if ecs.Log == nil {
		ecs.AppendParseError("Log level not found. Set to fallback")
		ecs.SetLogLevel(metalog.FallbackLoglevel)
	}
	if ecs.Timestamp == nil {
		ecs.AppendParseError("Timestamp not found. Set to fallback")
		ecs.SetTimeStamp(metalog.FallbackTimestamp)
	}
}
