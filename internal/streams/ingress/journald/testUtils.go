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
	testJournaldGrafanaLog = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_NAME":"grafana","COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"logfmt","COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION":"9.4.3.0","COM_HASHICORP_NOMAD_ALLOC_ID":"07ab1dac-04f7-fe70-b7d7-da2f0a488776","COM_HASHICORP_NOMAD_JOB_ID":"observability","COM_HASHICORP_NOMAD_JOB_NAME":"observability","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_NODE_ID":"0b854fe8-fa1a-1ec2-def2-914f1fae8dd7","COM_HASHICORP_NOMAD_NODE_NAME":"worker-01","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"grafana","COM_HASHICORP_NOMAD_TASK_NAME":"grafana","CONTAINER_ID":"bbe0c5a26c64","CONTAINER_ID_FULL":"bbe0c5a26c64d51214bcd91762284024748834e0e5ed93fbf09b6f4cb0eecb8b","CONTAINER_NAME":"grafana-07ab1dac-04f7-fe70-b7d7-da2f0a488776","CONTAINER_TAG":"bbe0c5a26c64","IMAGE_NAME":"registry.cloud.private/stack/observability/grafana:9.4.3.0","MAINTAINER":"Grafana Labs <hello@grafana.com>","PRIORITY":"6","SYSLOG_IDENTIFIER":"bbe0c5a26c64","_BOOT_ID":"283b57f0f4fc40698e9dd8959e071d04","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1064","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1678999436275171","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"987ee478f05b44588298516ca7a7b651","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"5396677057","__REALTIME_TIMESTAMP":"1678999436275223","host":"worker-01","message":"logger=live t=2023-03-16T20:43:56.274825539Z level=info msg=\"Initialized channel handler\" channel=grafana/dashboard/uid/KMg_v90Vz address=grafana/dashboard/uid/KMg_v90Vz","source_type":"journald","timestamp":"2023-03-16T20:43:56.275171Z"}`

	testJournaldTraefikInvalidLogFmt = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_NAME":"traefik","COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"logfmt","COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION":"2.9.8","COM_HASHICORP_NOMAD_ALLOC_ID":"40f22209-7d0f-899e-84ec-220fac8acdb4","COM_HASHICORP_NOMAD_JOB_ID":"ingress","COM_HASHICORP_NOMAD_JOB_NAME":"ingress","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_NODE_ID":"0b854fe8-fa1a-1ec2-def2-914f1fae8dd7","COM_HASHICORP_NOMAD_NODE_NAME":"worker-01","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"traefik","COM_HASHICORP_NOMAD_TASK_NAME":"traefik","CONTAINER_ID":"71f10a2bf888","CONTAINER_ID_FULL":"71f10a2bf88855997efdd4ffde0ef7e40af0fd9350546aef35e48bf0f8a5fce3","CONTAINER_NAME":"traefik-40f22209-7d0f-899e-84ec-220fac8acdb4","CONTAINER_TAG":"71f10a2bf888","IMAGE_NAME":"10.21.21.41:5000/traefik:v2.9.8","ORG_OPENCONTAINERS_IMAGE_DESCRIPTION":"A modern reverse-proxy","ORG_OPENCONTAINERS_IMAGE_DOCUMENTATION":"https://docs.traefik.io","ORG_OPENCONTAINERS_IMAGE_SOURCE":"https://github.com/traefik/traefik","ORG_OPENCONTAINERS_IMAGE_TITLE":"Traefik","ORG_OPENCONTAINERS_IMAGE_URL":"https://traefik.io","ORG_OPENCONTAINERS_IMAGE_VENDOR":"Traefik Labs","ORG_OPENCONTAINERS_IMAGE_VERSION":"v2.9.8","PRIORITY":"3","SYSLOG_IDENTIFIER":"71f10a2bf888","_BOOT_ID":"283b57f0f4fc40698e9dd8959e071d04","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1064","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1679006300545974","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"987ee478f05b44588298516ca7a7b651","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"12260948174","__REALTIME_TIMESTAMP":"1679006300546340","host":"worker-01","message":"2023/03/16 22:38:20 failed the request with status code 500","source_type":"journald","timestamp":"2023-03-16T22:38:20.545974Z"}`
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
