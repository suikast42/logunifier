package ecs

import (
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/patterns"
	"os"
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

func TestNativeEcs(t *testing.T) {
	log := `{"@timestamp":"2023-06-07T15:08:51.584+02:00","ecs":{"version":"1.3.0"},"log":{"level":"DEBUG","thread_name":"main","logger":"com.boxbay.wms.internal.test.curd.WmsCrudTest","origin":{"file":{"line":"56","name":"StartupInfoLogger.java"},"function":"logStarting"}},"service":{"name":"boxbay-wms-test","ephemeral_id":"cc6a891f-4642-485a-abd0-b13a230376e7"},"organization":{"name":"boxbay"},"host":{"hostname":"WAP130259","name":"WAP130259"},"message":"Running with Spring Boot v2.4.4, Spring v5.3.5"}`
	entry := EcsWrapper{}
	err := entry.FromJson([]byte(log))

	if err != nil {
		t.Error(err)
	}
}
