package dockerlogs

//
//import (
//	"encoding/json"
//	"github.com/nats-io/nats.go"
//	"github.com/suikast42/logunifier/internal/streams/ingress"
//	"github.com/suikast42/logunifier/pkg/model"
//	"github.com/suikast42/logunifier/pkg/patterns"
//	"google.golang.org/protobuf/types/known/timestamppb"
//	"strings"
//	"time"
//)
//
//type DockerToEcsConverter struct {
//}
//type IngessSubjectDockerLogs struct {
//	ContainerCreatedAt time.Time `json:"container_created_at"`
//	ContainerId        string    `json:"container_id"`
//	ContainerName      string    `json:"container_name"`
//	Host               string    `json:"host"`
//	Image              string    `json:"image"`
//	Label              struct {
//		ComHashicorpNomadAllocId       string `json:"com.hashicorp.nomad.alloc_id"`
//		ComHashicorpNomadJobId         string `json:"com.hashicorp.nomad.job_id"`
//		ComHashicorpNomadJobName       string `json:"com.hashicorp.nomad.job_name"`
//		ComHashicorpNomadNamespace     string `json:"com.hashicorp.nomad.namespace"`
//		ComHashicorpNomadNodeId        string `json:"com.hashicorp.nomad.node_id"`
//		ComHashicorpNomadNodeName      string `json:"com.hashicorp.nomad.node_name"`
//		ComHashicorpNomadTaskGroupName string `json:"com.hashicorp.nomad.task_group_name"`
//		ComHashicorpNomadTaskName      string `json:"com.hashicorp.nomad.task_name"`
//		OrgOpencontainersImageRevision string `json:"org.opencontainers.image.revision"`
//		OrgOpencontainersImageSource   string `json:"org.opencontainers.image.source"`
//		OrgOpencontainersImageTitle    string `json:"org.opencontainers.image.title"`
//	} `json:"label"`
//	Message    string    `json:"message"`
//	SourceType string    `json:"source_type"`
//	Stream     string    `json:"stream"`
//	Timestamp  time.Time `json:"timestamp"`
//}
//
//var containerToPattern = map[string]patterns.PatternKey{
//	"keycloak": patterns.KeyCloakPattern,
//	"nexus":    patterns.CommonUtcPatternWithCommaTsAndTz,
//}
//
//func (r *DockerToEcsConverter) Convert(msg *nats.Msg) *model.EcsLogEntry {
//	dockerLogEntry := IngessSubjectDockerLogs{}
//	err := json.Unmarshal(msg.Data, &dockerLogEntry)
//	if err != err {
//		return model.ToUnmarshalError(msg, err)
//	}
//	patternKey := dockerLogEntry.Label.ComHashicorpNomadTaskName
//	if patternKey == "" {
//		patternKey = dockerLogEntry.ContainerName
//	}
//	pattern, patternFound := containerToPattern[patternKey]
//
//	if !patternFound {
//		if strings.HasPrefix(patternKey, "connect-proxy-") {
//			containerToPattern[patternKey] = patterns.ConsulConnectPattern
//			pattern, patternFound = containerToPattern[patternKey]
//		} else if strings.HasSuffix(patternKey, "postgres") {
//			containerToPattern[patternKey] = patterns.ConsulConnectPattern
//			pattern, patternFound = containerToPattern[patternKey]
//		} else {
//			containerToPattern[patternKey] = patterns.CommonPattern
//			pattern, patternFound = containerToPattern[patternKey]
//		}
//	}
//
//	var parsed patterns.ParseResult
//	// A registered pattern found for message
//	def := patterns.ParseResult{
//		LogLevel:  model.LogLevel_unknown,
//		TimeStamp: dockerLogEntry.Timestamp,
//	}
//	parsed, err = patterns.Instance().ParseWitDefaults("IngessSubjectDockerLogs", patternKey, def, pattern, dockerLogEntry.Message)
//	if err != nil {
//		return model.ToUnmarshalError(msg, err)
//	}
//	return &model.EcsLogEntry{
//		Id:        model.UUID(),
//		Timestamp: timestamppb.New(parsed.TimeStamp),
//		Tags:      []string{dockerLogEntry.SourceType},
//		Log: &model.Log{
//			Level: parsed.LogLevel,
//		},
//		Message: dockerLogEntry.Message,
//		Container: &model.Container{
//			Id:        dockerLogEntry.ContainerId,
//			Name:      dockerLogEntry.ContainerName,
//			CreatedAt: timestamppb.New(dockerLogEntry.ContainerCreatedAt),
//			Image: &model.Container_Image{
//				Name: dockerLogEntry.Image,
//				Tag:  nil,
//			},
//			Labels: map[string]string{
//				ingress.IndexedContainerLabelStackName: dockerLogEntry.Label.ComHashicorpNomadJobName,
//				ingress.IndexedContainerLabelTaskGroup: dockerLogEntry.Label.ComHashicorpNomadTaskGroupName,
//				ingress.IndexedContainerLabelTask:      patternKey,
//				ingress.IndexedContainerLabelNamespace: dockerLogEntry.Label.ComHashicorpNomadNamespace,
//			},
//			Runtime: "",
//		},
//		Host: &model.Host{
//			Name: dockerLogEntry.Host,
//		},
//
//		Labels: map[string]string{
//			ingress.IndexedLabelIngress:     "vector-docker",
//			ingress.IndexedLabelUsedPattern: parsed.UsedPattern,
//		},
//	}
//}
