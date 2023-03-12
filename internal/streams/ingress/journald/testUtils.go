package journald

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
	"testing"
)

const (
	testJournaldContainerLog = `
{
"COM_HASHICORP_NOMAD_ALLOC_ID":"70a2254c-bed0-639b-86f3-3665b63ae732",
"COM_HASHICORP_NOMAD_JOB_ID":"observability",
"COM_HASHICORP_NOMAD_JOB_NAME":"observability",
"COM_HASHICORP_NOMAD_NAMESPACE":"default",
"COM_HASHICORP_NOMAD_NODE_ID":"bb71cdd9-ae46-835a-9ff4-547a45c4a4a3",
"COM_HASHICORP_NOMAD_NODE_NAME":"worker-01",
"COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"mimir",
"COM_HASHICORP_NOMAD_TASK_NAME":"mimir",
"CONTAINER_ID":"14f7262f9d64",
"CONTAINER_ID_FULL":"14f7262f9d64775fae3c78c8f7d20a7d46e1b321b0318f81089ba9e52e173853",
"CONTAINER_NAME":"mimir-70a2254c-bed0-639b-86f3-3665b63ae732",
"CONTAINER_TAG":"14f7262f9d64",
"IMAGE_NAME":"registry.cloud.private/grafana/mimir:2.6.0",
"ORG_OPENCONTAINERS_IMAGE_REVISION":"27698f3",
"ORG_OPENCONTAINERS_IMAGE_SOURCE":"https://github.com/grafana/mimir/tree/main/cmd/mimir",
"ORG_OPENCONTAINERS_IMAGE_TITLE":"mimir",
"PRIORITY":"3",
"SYSLOG_IDENTIFIER":"14f7262f9d64",
"_BOOT_ID":"2a9a0b466cc24dc6b24d8ddea42e83fa",
"_CAP_EFFECTIVE":"1ffffffffff",
"_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock",
"_COMM":"dockerd",
"_EXE":"/usr/bin/dockerd",
"_GID":"0",
"_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453",
"_PID":"1069",
"_SELINUX_CONTEXT":"unconfined\n",
"_SOURCE_REALTIME_TIMESTAMP":"1677232257226048",
"_SYSTEMD_CGROUP":"/system.slice/docker.service",
"_SYSTEMD_INVOCATION_ID":"2af5ab0aa3174940b91a125550a768fa",
"_SYSTEMD_SLICE":"system.slice",
"_SYSTEMD_UNIT":"docker.service",
"_TRANSPORT":"journal",
"_UID":"0",
"__MONOTONIC_TIMESTAMP":"6814551326",
"__REALTIME_TIMESTAMP":"1677232257226104",
"host":"worker-01",
"message":"ts=2023-02-24T09:50:57.225777448Z caller=logging.go:76 level=debug traceID=4738d36fd3f8854d msg=\"POST /api/v1/push (200) 8.088278ms\"",
"source_type":"journald",
"timestamp":"2023-02-24T09:50:57.226048Z"
}
`

	testJournaldDockerServiceLog = `
{"PRIORITY":"6","SYSLOG_FACILITY":"3","SYSLOG_IDENTIFIER":"dockerd","_BOOT_ID":"c974f369cb4a4fc0b8a1f19886f272e4","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1080","_SELINUX_CONTEXT":"unconfined\n","_STREAM_ID":"766e262320904a1384ac8bf6c017fd93","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"f69af9c9ac9243d79331c68ff81d2609","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"stdout","_UID":"0","__MONOTONIC_TIMESTAMP":"1525819321","__REALTIME_TIMESTAMP":"1678474432411076","host":"worker-01","message":"time=\"2023-03-10T18:53:52.411009960Z\" level=error msg=\"collecting stats for 7ba043edcb522051902c912313e7f53ef949f7435395db09d89a68047fbce8d3: Could not get container for 65a039e4e850a3299ea422e0d3ff0553ca97057ed9904c87c280848cae769294: No such container: 65a039e4e850a3299ea422e0d3ff0553ca97057ed9904c87c280848cae769294\"","source_type":"journald","timestamp":"2023-03-10T18:53:52.411076Z"}
`
)

func TestMetaLogFromJournalD(fromJson []byte, message string, t *testing.T) *model.MetaLog {

	converter := JournaldDToEcsConverter{}
	toMetaLog := converter.ConvertToMetaLog(&nats.Msg{
		Subject: "test",
		Header:  nil,
		Data:    fromJson,
		Sub:     nil,
	})
	if len(message) > 0 {
		toMetaLog.MetaLog.Message = message
	}
	return toMetaLog.MetaLog
}

func TestMetaLogFromJournalDFromConst(fromJson []byte, t *testing.T) *model.MetaLog {
	return TestMetaLogFromJournalD(fromJson, "", t)
}
