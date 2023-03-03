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
{
"PRIORITY":"6",
"SYSLOG_FACILITY":"3",
"SYSLOG_IDENTIFIER":"dockerd",
"_BOOT_ID":"2a9a0b466cc24dc6b24d8ddea42e83fa",
"_CAP_EFFECTIVE":"1ffffffffff",
"_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock",
"_COMM":"dockerd",
"_EXE":"/usr/bin/dockerd",
"_GID":"0",
"_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453",
"_PID":"1069",
"_SELINUX_CONTEXT":"unconfined\n",
"_STREAM_ID":"caa577ddf55a4eed9ba16490d7d572eb",
"_SYSTEMD_CGROUP":"/system.slice/docker.service",
"_SYSTEMD_INVOCATION_ID":"2af5ab0aa3174940b91a125550a768fa",
"_SYSTEMD_SLICE":"system.slice",
"_SYSTEMD_UNIT":"docker.service",
"_TRANSPORT":"stdout",
"_UID":"0",
"__MONOTONIC_TIMESTAMP":"6814378958",
"__REALTIME_TIMESTAMP":"1677232257053736",
"host":"worker-01",
"message":"time=\"2023-02-24T09:50:57.053689306Z\" level=error msg=\"collecting stats for 94cfdd08c026d0be0ae591868085c736210b588d1d8aacd33b1bba28e0a99c26: Could not get container for f326deec20f5f6f71d732995de0f82399934bd973728e6e70a37576d913c76cf: No such container: f326deec20f5f6f71d732995de0f82399934bd973728e6e70a37576d913c76cf\"",
"source_type":"journald",
"timestamp":"2023-02-24T09:50:57.053736Z"
}
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
	toMetaLog.MetaLog.Message = message
	return toMetaLog.MetaLog
}
