package dockerlogs

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

var unitToPattern map[string]patterns.PatternKey

type DockerToEcsConverter struct {
}
type IngessSubjectDockerLogs struct {
	ContainerCreatedAt time.Time `json:"container_created_at"`
	ContainerId        string    `json:"container_id"`
	ContainerName      string    `json:"container_name"`
	Host               string    `json:"host"`
	Image              string    `json:"image"`
	Label              struct {
		ComHashicorpNomadAllocId       string `json:"com.hashicorp.nomad.alloc_id"`
		ComHashicorpNomadJobId         string `json:"com.hashicorp.nomad.job_id"`
		ComHashicorpNomadJobName       string `json:"com.hashicorp.nomad.job_name"`
		ComHashicorpNomadNamespace     string `json:"com.hashicorp.nomad.namespace"`
		ComHashicorpNomadNodeId        string `json:"com.hashicorp.nomad.node_id"`
		ComHashicorpNomadNodeName      string `json:"com.hashicorp.nomad.node_name"`
		ComHashicorpNomadTaskGroupName string `json:"com.hashicorp.nomad.task_group_name"`
		ComHashicorpNomadTaskName      string `json:"com.hashicorp.nomad.task_name"`
		OrgOpencontainersImageRevision string `json:"org.opencontainers.image.revision"`
		OrgOpencontainersImageSource   string `json:"org.opencontainers.image.source"`
		OrgOpencontainersImageTitle    string `json:"org.opencontainers.image.title"`
	} `json:"label"`
	Message    string    `json:"message"`
	SourceType string    `json:"source_type"`
	Stream     string    `json:"stream"`
	Timestamp  time.Time `json:"timestamp"`
}

func init() {
	unitToPattern = make(map[string]patterns.PatternKey)

}

func (r *DockerToEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
	dockerLogEntry := IngessSubjectDockerLogs{}
	err := json.Unmarshal(msg.Data, &dockerLogEntry)
	if err != err {
		return model.ToUnmarshalError(msg, err)
	}
	return &model.EcsLogEntry{
		Id:        model.UUID(),
		Message:   dockerLogEntry.Message,
		Labels:    map[string]string{"ingress": "vector-docker"},
		Timestamp: timestamppb.New(dockerLogEntry.Timestamp),
	}
}
