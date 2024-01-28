package journald

import (
	"github.com/nats-io/nats.go"
	"github.com/suikast42/logunifier/pkg/model"
	"strings"
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
{"COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"logfmt","PRIORITY":"6","SYSLOG_FACILITY":"3","SYSLOG_IDENTIFIER":"dockerd","_BOOT_ID":"c974f369cb4a4fc0b8a1f19886f272e4","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1080","_SELINUX_CONTEXT":"unconfined\n","_STREAM_ID":"766e262320904a1384ac8bf6c017fd93","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"f69af9c9ac9243d79331c68ff81d2609","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"stdout","_UID":"0","__MONOTONIC_TIMESTAMP":"1525819321","__REALTIME_TIMESTAMP":"1678474432411076","host":"worker-01","message":"time=\"2023-03-10T18:53:52.411009960Z\" level=error msg=\"collecting stats for 7ba043edcb522051902c912313e7f53ef949f7435395db09d89a68047fbce8d3: Could not get container for 65a039e4e850a3299ea422e0d3ff0553ca97057ed9904c87c280848cae769294: No such container: 65a039e4e850a3299ea422e0d3ff0553ca97057ed9904c87c280848cae769294\"","source_type":"journald","timestamp":"2023-03-10T18:53:52.411076Z"}
`
	testJournaldGrafanaLog = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_NAME":"grafana","COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"logfmt","COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION":"9.4.3.0","COM_HASHICORP_NOMAD_ALLOC_ID":"07ab1dac-04f7-fe70-b7d7-da2f0a488776","COM_HASHICORP_NOMAD_JOB_ID":"observability","COM_HASHICORP_NOMAD_JOB_NAME":"observability","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_NODE_ID":"0b854fe8-fa1a-1ec2-def2-914f1fae8dd7","COM_HASHICORP_NOMAD_NODE_NAME":"worker-01","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"grafana","COM_HASHICORP_NOMAD_TASK_NAME":"grafana","CONTAINER_ID":"bbe0c5a26c64","CONTAINER_ID_FULL":"bbe0c5a26c64d51214bcd91762284024748834e0e5ed93fbf09b6f4cb0eecb8b","CONTAINER_NAME":"grafana-07ab1dac-04f7-fe70-b7d7-da2f0a488776","CONTAINER_TAG":"bbe0c5a26c64","IMAGE_NAME":"registry.cloud.private/stack/observability/grafana:9.4.3.0","MAINTAINER":"Grafana Labs <hello@grafana.com>","PRIORITY":"6","SYSLOG_IDENTIFIER":"bbe0c5a26c64","_BOOT_ID":"283b57f0f4fc40698e9dd8959e071d04","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1064","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1678999436275171","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"987ee478f05b44588298516ca7a7b651","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"5396677057","__REALTIME_TIMESTAMP":"1678999436275223","host":"worker-01","message":"logger=live t=2023-03-16T20:43:56.274825539Z level=info msg=\"Initialized channel handler\" channel=grafana/dashboard/uid/KMg_v90Vz address=grafana/dashboard/uid/KMg_v90Vz","source_type":"journald","timestamp":"2023-03-16T20:43:56.275171Z"}`

	testJournaldTraefikInvalidLogFmt = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_NAME":"traefik","COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"logfmt","COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION":"2.9.8","COM_HASHICORP_NOMAD_ALLOC_ID":"40f22209-7d0f-899e-84ec-220fac8acdb4","COM_HASHICORP_NOMAD_JOB_ID":"ingress","COM_HASHICORP_NOMAD_JOB_NAME":"ingress","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_NODE_ID":"0b854fe8-fa1a-1ec2-def2-914f1fae8dd7","COM_HASHICORP_NOMAD_NODE_NAME":"worker-01","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"traefik","COM_HASHICORP_NOMAD_TASK_NAME":"traefik","CONTAINER_ID":"bf41c622fca3","CONTAINER_ID_FULL":"bf41c622fca3d63578ba1f4ab253da32259cd0dbaf9eeea72984e5d92def2047","CONTAINER_NAME":"traefik-40f22209-7d0f-899e-84ec-220fac8acdb4","CONTAINER_TAG":"bf41c622fca3","IMAGE_NAME":"10.21.21.41:5000/traefik:v2.9.8","ORG_OPENCONTAINERS_IMAGE_DESCRIPTION":"A modern reverse-proxy","ORG_OPENCONTAINERS_IMAGE_DOCUMENTATION":"https://docs.traefik.io","ORG_OPENCONTAINERS_IMAGE_SOURCE":"https://github.com/traefik/traefik","ORG_OPENCONTAINERS_IMAGE_TITLE":"Traefik","ORG_OPENCONTAINERS_IMAGE_URL":"https://traefik.io","ORG_OPENCONTAINERS_IMAGE_VENDOR":"Traefik Labs","ORG_OPENCONTAINERS_IMAGE_VERSION":"v2.9.8","PRIORITY":"3","SYSLOG_IDENTIFIER":"bf41c622fca3","_BOOT_ID":"a29e621a1bd44288b18425f925972679","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1068","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1679151644250632","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"da9b1318a7264daea60661fcead91e73","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"209848294","__REALTIME_TIMESTAMP":"1679151644250686","host":"worker-01","message":"2023/03/18 15:00:44 failed to send the request: Post \"http://tempo-zipkin.service.consul:9411/api/v2/spans\": dial tcp: lookup tempo-zipkin.service.consul on 10.21.21.42:53: no such host","source_type":"journald","timestamp":"2023-03-18T15:00:44.250632Z"}`
	testJournaldNomadLog             = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"tsLevelMsg", "PRIORITY":"6","SYSLOG_FACILITY":"3","SYSLOG_IDENTIFIER":"nomad","_BOOT_ID":"4cc34f56f72249e395a9e09606fc4656","_CAP_EFFECTIVE":"0","_CMDLINE":"/usr/local/bin/nomad agent -config /etc/nomad.d","_COMM":"nomad","_EXE":"/usr/local/bin/nomad","_GID":"1004","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"862","_SELINUX_CONTEXT":"unconfined\n","_STREAM_ID":"933425e2609f412cba0967953d5f3005","_SYSTEMD_CGROUP":"/system.slice/nomad.service","_SYSTEMD_INVOCATION_ID":"adb44878c2114917a4145cc9516ca640","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"nomad.service","_TRANSPORT":"stdout","_UID":"997","__MONOTONIC_TIMESTAMP":"2461545048","__REALTIME_TIMESTAMP":"1679324805057689","host":"master-01","message":"    2023-03-20T15:06:45.057Z [DEBUG] nomad: memberlist: Stream connection from=127.0.0.1:48046","source_type":"journald","timestamp":"2023-03-20T15:06:45.057689Z"}`
	testJournaldConsulConnect        = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"envoy","COM_HASHICORP_NOMAD_ALLOC_ID":"9088f933-d4fd-f1ef-91b9-e12236ccbe79","COM_HASHICORP_NOMAD_JOB_ID":"observability","COM_HASHICORP_NOMAD_JOB_NAME":"observability","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_NODE_ID":"0b854fe8-fa1a-1ec2-def2-914f1fae8dd7","COM_HASHICORP_NOMAD_NODE_NAME":"worker-01","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"grafana","COM_HASHICORP_NOMAD_TASK_NAME":"connect-proxy-grafana","CONTAINER_ID":"313872e42969","CONTAINER_ID_FULL":"313872e4296925b6d853c14ed787e866c551993fffd3ec131dcfc1bd36a892c4","CONTAINER_NAME":"connect-proxy-grafana-9088f933-d4fd-f1ef-91b9-e12236ccbe79","CONTAINER_TAG":"313872e42969","IMAGE_NAME":"envoyproxy/envoy:v1.25.1","ORG_OPENCONTAINERS_IMAGE_REF_NAME":"ubuntu","ORG_OPENCONTAINERS_IMAGE_VERSION":"20.04","PRIORITY":"3","SYSLOG_IDENTIFIER":"313872e42969","_BOOT_ID":"1b02f855e61b4525837e8c31fe47031c","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1089","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1680179600424739","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"1f2ef6f4fd0448798315845a376ad7e8","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"70387876311","__REALTIME_TIMESTAMP":"1680179600424788","host":"worker-01","message":"[2023-03-30 12:33:20.424][15][debug][rbac] [source/extensions/filters/network/rbac/rbac_filter.cc:90] checking connection: requestedServerName: , sourceIP: 10.21.21.42:46746, directRemoteIP: 10.21.21.42:46746,remoteIP: 10.21.21.42:46746, localAddress: 172.26.68.105:26417, ssl: uriSanPeerCertificate: spiffe://1eb6deee-8554-9c9a-ddca-cefc6d80c81e.consul/ns/default/dc/nomadder1/svc/traefik, dnsSanPeerCertificate: , subjectPeerCertificate: , dynamicMetadata: ","source_type":"journald","timestamp":"2023-03-30T12:33:20.424739Z"}`

	testJournaldLogunifier        = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_NAME":"logunifier","COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"tslevelmsg","COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION":"0.1.0","COM_GITHUB_LOGUNIFIER_APPLICATION_STRIP_ANSI":"true","COM_HASHICORP_NOMAD_ALLOC_ID":"bab93287-6e17-1849-22cc-7449612bf642","COM_HASHICORP_NOMAD_JOB_ID":"observability","COM_HASHICORP_NOMAD_JOB_NAME":"observability","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_NODE_ID":"0b854fe8-fa1a-1ec2-def2-914f1fae8dd7","COM_HASHICORP_NOMAD_NODE_NAME":"worker-01","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"logunifier","COM_HASHICORP_NOMAD_TASK_NAME":"logunifier","CONTAINER_ID":"44d89924c110","CONTAINER_ID_FULL":"44d89924c110bd70e155ff7a9b14d2f995fca6738f2d59ab568f67f00bdc47d1","CONTAINER_NAME":"logunifier-bab93287-6e17-1849-22cc-7449612bf642","CONTAINER_TAG":"44d89924c110","IMAGE_NAME":"registry.cloud.private/suikast42/logunifier:0.1.0","PRIORITY":"6","SYSLOG_IDENTIFIER":"44d89924c110","_BOOT_ID":"65021dc0363148db80d5c06753cacc07","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1060","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1680207232774470","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"de15cfebb084499e8f296f638a37cbd7","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"2363870502","__REALTIME_TIMESTAMP":"1680207232774527","host":"worker-01","message":"\u001b[90m2023-03-30T20:13:52.774125Z\u001b[0m \u001b[33mDBG\u001b[0m Nothing to validate after 10s ","source_type":"journald","timestamp":"2023-03-30T20:13:52.774470Z"}`
	testJournaldInvalidTsLevelMsg = `{"ARCHITECTURE":"x86_64","BUILD_DATE":"2023-03-22T10:50:14","COM_DOCKER_COMPOSE_CONFIG_HASH":"2e39f0334b122364ec3f21517385e457ccfdf42c4fa580a42c4462901ec38600","COM_DOCKER_COMPOSE_CONTAINER_NUMBER":"1","COM_DOCKER_COMPOSE_DEPENDS_ON":"","COM_DOCKER_COMPOSE_IMAGE":"sha256:d73029e32cc349d58285684f1d656375381c7e1e5e8a744d91f32b50e72e91fc","COM_DOCKER_COMPOSE_ONEOFF":"False","COM_DOCKER_COMPOSE_PROJECT":"core","COM_DOCKER_COMPOSE_PROJECT_CONFIG_FILES":"/opt/nomadjobs/core/docker-compose.yml","COM_DOCKER_COMPOSE_PROJECT_WORKING_DIR":"/opt/nomadjobs/core","COM_DOCKER_COMPOSE_REPLACE":"73430a00aab79d7d702841e145e9b05e835e67be29475d010c99ab67964224bb","COM_DOCKER_COMPOSE_SERVICE":"nexus","COM_DOCKER_COMPOSE_VERSION":"2.16.0","COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"tslevelmsg","COM_HASHICORP_NOMAD_JOB_NAME":"nexus","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"nexus","COM_HASHICORP_NOMAD_TASK_NAME":"nexus","COM_REDHAT_COMPONENT":"ubi8-minimal-container","COM_REDHAT_LICENSE_TERMS":"https://www.redhat.com/en/about/red-hat-end-user-license-agreements#UBI","COM_SONATYPE_LICENSE":"Apache License, Version 2.0","COM_SONATYPE_NAME":"Nexus Repository Manager base image","CONTAINER_ID":"171162c4fbe4","CONTAINER_ID_FULL":"171162c4fbe48e379b858ca38c608834584b2b637e9e56858b4ba82b80a5ef03","CONTAINER_NAME":"nexus","CONTAINER_TAG":"171162c4fbe4","DESCRIPTION":"The Nexus Repository Manager server           with universal support for popular component formats.","DISTRIBUTION_SCOPE":"public","IMAGE_NAME":"sonatype/nexus3:3.50.0","IO_BUILDAH_VERSION":"1.27.3","IO_K8S_DESCRIPTION":"The Nexus Repository Manager server           with universal support for popular component formats.","IO_K8S_DISPLAY_NAME":"Nexus Repository Manager","IO_OPENSHIFT_EXPOSE_SERVICES":"8081:8081","IO_OPENSHIFT_TAGS":"Sonatype,Nexus,Repository Manager","MAINTAINER":"Sonatype <support@sonatype.com>","NAME":"Nexus Repository Manager","PRIORITY":"6","RELEASE":"3.50.0","RUN":"docker run -d --name NAME           -p 8081:8081           IMAGE","STOP":"docker stop NAME","SUMMARY":"The Nexus Repository Manager server           with universal support for popular component formats.","SYSLOG_IDENTIFIER":"171162c4fbe4","URL":"https://sonatype.com","VCS_REF":"146fdafc2595e26f5f9c1b9a2b3f36bbca8237e4","VCS_TYPE":"git","VENDOR":"Sonatype","VERSION":"3.50.0-01","_BOOT_ID":"b6e62de3cc7e4243b04dcf129f3e4b84","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1031","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1680045082179163","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"fdd3e18442c54bc183018cea9af633e4","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"10328048728","__REALTIME_TIMESTAMP":"1680045082179850","host":"master-01","message":"Invalid message","source_type":"journald","timestamp":"2023-03-28T23:11:22.179163Z"}`

	testNatviceEcs  = `{"@timestamp":"2023-06-07T15:08:51.584+02:00","ecs":{"version":"1.3.0"},"log":{"level":"DEBUG","thread_name":"main","logger":"com.boxbay.wms.internal.test.curd.WmsCrudTest","origin":{"file":{"line":"56","name":"StartupInfoLogger.java"},"function":"logStarting"}},"service":{"name":"boxbay-wms-test","ephemeral_id":"cc6a891f-4642-485a-abd0-b13a230376e7"},"organization":{"name":"boxbay"},"host":{"hostname":"WAP130259","name":"WAP130259"},"message":"Running with Spring Boot v2.4.4, Spring v5.3.5"}`
	testJournaldEcs = `{"COM_GITHUB_LOGUNIFIER_APPLICATION_NAME":"grafana","COM_GITHUB_LOGUNIFIER_APPLICATION_PATTERN_KEY":"ecs","COM_GITHUB_LOGUNIFIER_APPLICATION_VERSION":"9.4.3.0","COM_HASHICORP_NOMAD_ALLOC_ID":"07ab1dac-04f7-fe70-b7d7-da2f0a488776","COM_HASHICORP_NOMAD_JOB_ID":"observability","COM_HASHICORP_NOMAD_JOB_NAME":"observability","COM_HASHICORP_NOMAD_NAMESPACE":"default","COM_HASHICORP_NOMAD_NODE_ID":"0b854fe8-fa1a-1ec2-def2-914f1fae8dd7","COM_HASHICORP_NOMAD_NODE_NAME":"worker-01","COM_HASHICORP_NOMAD_TASK_GROUP_NAME":"grafana","COM_HASHICORP_NOMAD_TASK_NAME":"grafana","CONTAINER_ID":"bbe0c5a26c64","CONTAINER_ID_FULL":"bbe0c5a26c64d51214bcd91762284024748834e0e5ed93fbf09b6f4cb0eecb8b","CONTAINER_NAME":"grafana-07ab1dac-04f7-fe70-b7d7-da2f0a488776","CONTAINER_TAG":"bbe0c5a26c64","IMAGE_NAME":"registry.cloud.private/stack/observability/grafana:9.4.3.0","MAINTAINER":"Grafana Labs <hello@grafana.com>","PRIORITY":"6","SYSLOG_IDENTIFIER":"bbe0c5a26c64","_BOOT_ID":"283b57f0f4fc40698e9dd8959e071d04","_CAP_EFFECTIVE":"1ffffffffff","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","_COMM":"dockerd","_EXE":"/usr/bin/dockerd","_GID":"0","_MACHINE_ID":"ceacb99587e34bcc840bc7a7cc0d4453","_PID":"1064","_SELINUX_CONTEXT":"unconfined\n","_SOURCE_REALTIME_TIMESTAMP":"1678999436275171","_SYSTEMD_CGROUP":"/system.slice/docker.service","_SYSTEMD_INVOCATION_ID":"987ee478f05b44588298516ca7a7b651","_SYSTEMD_SLICE":"system.slice","_SYSTEMD_UNIT":"docker.service","_TRANSPORT":"journal","_UID":"0","__MONOTONIC_TIMESTAMP":"5396677057","__REALTIME_TIMESTAMP":"1678999436275223","host":"worker-01","message":"##MSG##","source_type":"journald","timestamp":"2023-03-16T20:43:56.275171Z"}`
)

func TestMetaLogFromJournalD(fromJson []byte, message string, t *testing.T) *model.MetaLog {

	converter := JournaldDToEcsConverter{}
	if string(fromJson) == testJournaldEcs {
		embeddedJsonMessage := strings.ReplaceAll(testNatviceEcs, "\"", "\\\"")
		fromJson = []byte(strings.ReplaceAll(testJournaldEcs, "##MSG##", embeddedJsonMessage))
	}
	toMetaLog := converter.ConvertToMetaLog(&nats.Msg{
		Subject: "test",
		Header:  nil,
		Data:    fromJson,
		Sub:     nil,
	})
	if len(message) > 0 {
		toMetaLog.MetaLog.EcsLogEntry.Message = message
		toMetaLog.MetaLog.RawMessage = message
	}

	return toMetaLog.MetaLog
}

func TestMetaLogFromJournalDFromConst(fromJson []byte, t *testing.T) *model.MetaLog {
	return TestMetaLogFromJournalD(fromJson, "", t)
}
