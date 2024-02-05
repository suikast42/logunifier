package ecs

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/patterns"
	"os"
	"strings"
	"testing"
)

var (
	patternfactory *patterns.PatternFactory
	logger         zerolog.Logger
)

func init() {
	logger = config.Logger()
	_, err := patterns.Initialize()
	if err != nil {
		logger.Error().Err(err).Stack().Msg("Can't initialize pattern factory")
		os.Exit(1)
	}
	patternfactory = patterns.Instance()
}

func TestEcsVectorIngress(t *testing.T) {
	log := `{
  "@timestamp": "2022-08-05T11:33:20.672+02:00",
  "ecs": {
    "version": "1.3.0"
  },
  "log": {
    "level": "WARN",
    "thread_name": "ServerService Thread Pool -- 62",
    "logger": "org.hibernate.tool.schema.internal.ExceptionHandlerLoggedImpl",
    "origin": {
      "file": {
        "line": "27",
        "name": "ExceptionHandlerLoggedImpl.java"
      },
      "function": "handleException"
    }
  },
  "service": {
    "name": "wms-kemp",
    "ephemeral_id": "c4acf556-6126-4b79-8c7b-931319e156e1"
  },
  "organization": {
    "name": "kemp"
  },
  "host": {
    "hostname": "WAP130259",
    "name": "WAP130259"
  },
  "message": "GenerationTarget encountered exception accepting command : Error executing DDL \"alter table USERROLE_USER add constraint USER_ID foreign key (USERROLE_ID) references WMSUSERROLE\" via JDBC Statement",
  "error": {
    "stack_trace": "org.hibernate.tool.schema.spi.CommandAcceptanceException: Error executing DDL \"alter table USERROLE_USER add constraint USER_ID foreign key (USERROLE_ID) references WMSUSERROLE\" via JDBC Statement\n\tat org.hibernate.tool.schema.internal.exec.GenerationTargetToDatabase.accept(GenerationTargetToDatabase.java:67)\n\tat org.hibernate.tool.schema.internal.SchemaCreatorImpl.applySqlString(SchemaCreatorImpl.java:440)\n\tat org.hibernate.tool.schema.internal.SchemaCreatorImpl.applySqlStrings(SchemaCreatorImpl.java:424)\n\tat org.hibernate.tool.schema.internal.SchemaCreatorImpl.createFromMetadata(SchemaCreatorImpl.java:375)\n\tat org.hibernate.tool.schema.internal.SchemaCreatorImpl.performCreation(SchemaCreatorImpl.java:166)\n\tat org.hibernate.tool.schema.internal.SchemaCreatorImpl.doCreation(SchemaCreatorImpl.java:135)\n\tat org.hibernate.tool.schema.internal.SchemaCreatorImpl.doCreation(SchemaCreatorImpl.java:121)\n\tat org.hibernate.tool.schema.spi.SchemaManagementToolCoordinator.performDatabaseAction(SchemaManagementToolCoordinator.java:155)\n\tat org.hibernate.tool.schema.spi.SchemaManagementToolCoordinator.process(SchemaManagementToolCoordinator.java:72)\n\tat org.hibernate.internal.SessionFactoryImpl.<init>(SessionFactoryImpl.java:310)\n\tat org.hibernate.boot.internal.SessionFactoryBuilderImpl.build(SessionFactoryBuilderImpl.java:467)\n\tat org.hibernate.jpa.boot.internal.EntityManagerFactoryBuilderImpl.build(EntityManagerFactoryBuilderImpl.java:939)\n\tat org.jboss.as.jpa.hibernate5.TwoPhaseBootstrapImpl.build(TwoPhaseBootstrapImpl.java:44)\n\tat org.jboss.as.jpa.service.PersistenceUnitServiceImpl$1$1.run(PersistenceUnitServiceImpl.java:170)\n\tat org.jboss.as.jpa.service.PersistenceUnitServiceImpl$1$1.run(PersistenceUnitServiceImpl.java:128)\n\tat org.wildfly.security.manager.WildFlySecurityManager.doChecked(WildFlySecurityManager.java:649)\n\tat org.jboss.as.jpa.service.PersistenceUnitServiceImpl$1.run(PersistenceUnitServiceImpl.java:212)\n\tat org.jboss.threads.ContextClassLoaderSavingRunnable.run(ContextClassLoaderSavingRunnable.java:35)\n\tat org.jboss.threads.EnhancedQueueExecutor.safeRun(EnhancedQueueExecutor.java:1982)\n\tat org.jboss.threads.EnhancedQueueExecutor$ThreadBody.doRunTask(EnhancedQueueExecutor.java:1486)\n\tat org.jboss.threads.EnhancedQueueExecutor$ThreadBody.run(EnhancedQueueExecutor.java:1377)\n\tat java.lang.Thread.run(Thread.java:748)\n\tat org.jboss.threads.JBossThread.run(JBossThread.java:485)\nCaused by: com.microsoft.sqlserver.jdbc.SQLServerException: There is already an object named 'USER_ID' in the database.\n\tat com.microsoft.sqlserver.jdbc.SQLServerException.makeFromDatabaseError(SQLServerException.java:254)\n\tat com.microsoft.sqlserver.jdbc.SQLServerStatement.getNextResult(SQLServerStatement.java:1608)\n\tat com.microsoft.sqlserver.jdbc.SQLServerStatement.doExecuteStatement(SQLServerStatement.java:859)\n\tat com.microsoft.sqlserver.jdbc.SQLServerStatement$StmtExecCmd.doExecute(SQLServerStatement.java:759)\n\tat com.microsoft.sqlserver.jdbc.TDSCommand.execute(IOBuffer.java:7240)\n\tat com.microsoft.sqlserver.jdbc.SQLServerConnection.executeCommand(SQLServerConnection.java:2869)\n\tat com.microsoft.sqlserver.jdbc.SQLServerStatement.executeCommand(SQLServerStatement.java:243)\n\tat com.microsoft.sqlserver.jdbc.SQLServerStatement.executeStatement(SQLServerStatement.java:218)\n\tat com.microsoft.sqlserver.jdbc.SQLServerStatement.execute(SQLServerStatement.java:739)\n\tat org.jboss.jca.adapters.jdbc.WrappedStatement.execute(WrappedStatement.java:198)\n\tat org.hibernate.tool.schema.internal.exec.GenerationTargetToDatabase.accept(GenerationTargetToDatabase.java:54)\n\t... 22 more",
    "message": "Error executing DDL \"alter table USERROLE_USER add constraint USER_ID foreign key (USERROLE_ID) references WMSUSERROLE\" via JDBC Statement",
    "type": "org.hibernate.tool.schema.spi.CommandAcceptanceException"
  },
  "tags": []
}`
	entry := EcsWrapper{}
	err := entry.FromJson([]byte(log))

	if err != nil {
		t.Error(err)
	}

}

// With log line as string
func TestNativeEcs1(t *testing.T) {
	log := `{"@timestamp":"2024-01-06T00:18:29.089+01:00","ecs":{"version":"1.3.0"},"log":{"level":"DEBUG","thread_name":"main","logger":"com.boxbay.wms.internal.test.curd.WmsCrudTest","origin":{"file":{"line":"56","name":"StartupInfoLogger.java"},"function":"logStarting"}},"service":{"name":"boxbay-wms-test","ephemeral_id":"cc6a891f-4642-485a-abd0-b13a230376e7"},"organization":{"name":"boxbay"},"host":{"hostname":"WAP130259","name":"WAP130259"},"message":"Running with Spring Boot v2.4.4, Spring v5.3.5"}`
	entry := EcsWrapper{}
	err := entry.FromJson([]byte(log))

	if err != nil {
		t.Error(err)
	}

	if "2024-01-05 23:18:29.089 +0000 UTC" != entry.Timestamp.AsTime().String() {
		t.Errorf("TS does not match %s", entry.Timestamp.AsTime().String())
	}
}

// With log line as int
func TestNativeEcs2(t *testing.T) {
	log := `{"@timestamp":"2024-01-06T00:18:29.089+01:00","ecs":{"version":"1.3.0"},"log":{"level":"DEBUG","thread_name":"scheduling-1","logger":"com.boxbay.emulation.stc.integration.StcEmulationIntegrationCommunication","origin":{"file":{"line":251,"name":"StcEmulationIntegrationCommunication.java"},"function":"handle"}},"service":{"name":"stc-emulation","ephemeral_id":"3622a316-da26-47c5-805c-16a9643819b8"},"organization":{"name":"boxbay"},"host":{"hostname":"WAP153441","name":"WAP153441"},"message":"Position: X: 78794  Y: 199  Z: 0  R1: 100  R2: 3165"}`
	entry := EcsWrapper{}
	err := entry.FromJson([]byte(log))

	if err != nil {
		t.Error(err)
	}

	if "2024-01-05 23:18:29.089 +0000 UTC" != entry.Timestamp.AsTime().String() {
		t.Errorf("TS does not match %s", entry.Timestamp.AsTime().String())
	}

}

func TestNativeEcs3(t *testing.T) {
	log := `
{
  "@timestamp": "2024-01-29T15:21:52.136+01:00",
  "ecs": {
    "version": "1.3.0"
  },
  "log": {
    "level": "DEBUG",
    "thread_name": "Thread-809 (ActiveMQ-client-global-threads)",
    "logger": "eu.amova.assan.server.routing.ASSANHbsManagerOPCache",
    "origin": {
      "file": {
        "line": "1719",
        "name": "AbstractAreaManagerOPCache.java"
      },
      "function": "logEnterMethod"
    }
  },
  "service": {
    "name": "wms-assan",
    "ephemeral_id": "77764cbf-f214-44e1-9db7-a5b67f207282"
  },
  "organization": {
    "name": "assan"
  },
  "host": {
    "hostname": "WAP153441",
    "name": "WAP153441"
  },
  "message": "ENTER Method : getAreaManagerByDevice",
  "tags": []
}
`
	ecs := model.EcsLogEntry{}
	err := ecs.FromJson([]byte(log))

	if err != nil {
		t.Error(err)
	}

	if "2024-01-29 14:21:52.136 +0000 UTC" != ecs.Timestamp.AsTime().String() {
		t.Errorf("TS does not match %s", ecs.Timestamp.AsTime().String())
	}

	if "wms-assan" != ecs.Service.Name {
		t.Errorf("ecs.Service.Name does not match %s", ecs.Service.Name)
	}
	//if "1.3.0" != ecs.Version.GetVersion() {
	//	t.Errorf("Version does not match %s", ecs.Version)
	//}
}

func TestNativeEcsApm(t *testing.T) {
	log := `
{
  "environment": {
    "name": "dev"
  },
  "trace": {
   "trace": {
    "id": "8795cd88e90fbfce1e35acfcfaeada5e"
  },
  "span": {
    "id": "8795cd88e90fbfce1e35acfcfaeada5e"
  }
  },
  "@timestamp": "2024-02-05T13:22:25.483+01:00",
  "ecs": {
    "version": "1.3.0"
  },
  "log": {
    "level": "INFO",
    "thread_name": "ServerService Thread Pool -- 48",
    "logger": "pas.core.server.initalize.WMSInitalizer",
    "origin": {
      "file": {
        "line": "66",
        "name": "WMSInitalizer.java"
      },
      "function": "initCompos"
    }
  },
  "service": {
    "stack": "wms",
    "name": "wms-assan",
    "namespace": "default",
    "ephemeral_id": "5c5160cc-b79e-40cc-b8e5-e4a5f17adf86",
    "type": "external",
    "version": "1.0.0-SNAPSHOT"
  },
  "organization": {
    "name": "assan"
  },
  "host": {
    "hostname": "WAP153441",
    "name": "WAP153441"
  },
  "message": "Starting the initialisation phase for ASSAN",
  "tags": []
}
`
	ecs := model.EcsLogEntry{}
	err := ecs.FromJson([]byte(log))

	if err != nil {
		t.Error(err)
	}

	if !ecs.IsTraceIdSet() {
		t.Error("Trace id not set ")
	}

	if !ecs.IsSpanIdSet() {
		t.Error("Span id not set ")
	}

	if ecs.IsTraceIdSet() && !strings.EqualFold(ecs.Trace.Trace.Id, "8795cd88e90fbfce1e35acfcfaeada5e") {
		t.Errorf("expect trace id  %s  but was %s", "8795cd88e90fbfce1e35acfcfaeada5e", ecs.Trace.Trace.Id)
	}

	if ecs.IsSpanIdSet() && !strings.EqualFold(ecs.Trace.Span.Id, "8795cd88e90fbfce1e35acfcfaeada5e") {
		t.Errorf("expect span id  %s  but was %s", "8795cd88e90fbfce1e35acfcfaeada5e", ecs.Trace.Span.Id)
	}
	//if "1.3.0" != ecs.Version.GetVersion() {
	//	t.Errorf("Version does not match %s", ecs.Version)
	//}
}
