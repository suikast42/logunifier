package config

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// region type Config
type Config struct {
	ingressNatsJournald  string
	ingressNatsNativeEcs string
	//ingressNatsDocker   string
	ingresSubjectTest string
	natsServers       []string
	lokiServers       []string
	loglevel          string
	egressSubjectEcs  string
	ackTimeoutS       int
	pingLog           bool
}

func (c Config) AckTimeoutS() int {
	return c.ackTimeoutS
}

func (c Config) String() string {
	return fmt.Sprintf(`Config:
ingressNatsJournald: %v,
natsServers: %v,
loglevel: %v,
`, c.ingressNatsJournald, c.natsServers, c.loglevel)
}

func (c Config) IngressNatsJournald() string {
	return c.ingressNatsJournald
}

func (c Config) IngressNatsNativeEcs() string {
	return c.ingressNatsNativeEcs
}

func (c Config) PingLog() bool {
	return c.pingLog
}

//func (c Config) IngressNatsDocker() string {
//	return c.ingressNatsDocker
//}

func (c Config) IngresNatsTest() string {
	return c.ingresSubjectTest
}
func (c Config) NatsServers() string {
	return strings.Join(c.natsServers, ",")
}

func (c Config) LokiServers() []string {
	return c.lokiServers
}

func (c Config) Loglevel() string {
	return c.loglevel
}

func (c Config) EgressSubjectEcs() string {
	return c.egressSubjectEcs
}

//endregion

// region enums
type IngressSubject string

//endregion

//region builder

type ConfigBuilder struct {
	cfg *Config
}

func (r *ConfigBuilder) withNatsServer(server string) *ConfigBuilder {
	for _, s := range r.cfg.natsServers {
		if strings.Compare(s, server) == 0 {
			return r
		}
	}
	r.cfg.natsServers = append(r.cfg.natsServers, server)
	return r
}

func (r *ConfigBuilder) withLokiServers(server string) *ConfigBuilder {
	for _, s := range r.cfg.lokiServers {
		if strings.Compare(s, server) == 0 {
			return r
		}
	}
	r.cfg.lokiServers = append(r.cfg.lokiServers, server)
	return r
}

func (r *ConfigBuilder) withLogLevel(loglevel *string) *ConfigBuilder {
	r.cfg.loglevel = *loglevel
	return r
}

func (r *ConfigBuilder) withAckTimeout(ackTimeout *int) *ConfigBuilder {
	r.cfg.ackTimeoutS = *ackTimeout
	return r
}

func (r *ConfigBuilder) withEgressSubjectEcs(egressSubjectEcs *string) *ConfigBuilder {
	r.cfg.egressSubjectEcs = *egressSubjectEcs
	return r
}

func (r *ConfigBuilder) withIngressSubjectJournald(ingressNatsJournald *string) *ConfigBuilder {
	r.cfg.ingressNatsJournald = *ingressNatsJournald
	return r
}

func (r *ConfigBuilder) withIngressSubjectNativeEcs(ingressNatsNativeEcs *string) *ConfigBuilder {
	r.cfg.ingressNatsNativeEcs = *ingressNatsNativeEcs
	return r
}
func (r *ConfigBuilder) withIngresSubjectTest(ingresSubjectTest *string) *ConfigBuilder {
	r.cfg.ingresSubjectTest = *ingresSubjectTest
	return r
}

func (r *ConfigBuilder) withPingLog(pingLog *bool) *ConfigBuilder {
	r.cfg.pingLog = *pingLog
	return r
}

//	func (r *ConfigBuilder) withIngressSubjectDocker(ingressNatsDocker *string) *ConfigBuilder {
//		r.cfg.ingressNatsDocker = *ingressNatsDocker
//		return r
//	}
func (r *ConfigBuilder) build() *Config {
	lock.Lock()
	defer lock.Unlock()
	instance = r.cfg
	config, _ := Instance()
	return config
}

func newBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		cfg: &Config{},
	}
}

//endregion

// region singleton
var instance *Config
var lock = &sync.Mutex{}

func Instance() (*Config, error) {
	if instance == nil {
		return nil, errors.New("no config parsed. Use the builder at first")
	}

	return instance, nil
}

//endregion
