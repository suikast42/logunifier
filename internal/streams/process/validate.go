package process

import (
	"github.com/suikast42/logunifier/pkg/model"
)

func Validate(ecs *model.EcsLogEntry) {
	if !ecs.IsJobNameSet() {
		ecs.AppendParseError("Job name is empty ")
		ecs.SetJobName("Empty")
	}
	if !ecs.IsJobTypeSet() {
		ecs.AppendParseError("Job type is empty")
		ecs.SetJobType("Empty")
	}
}
