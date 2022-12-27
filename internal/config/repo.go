package config

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// region type Config
type Config struct {
	ingressNatsJournald string
	ingresSubjectTest   string
	natsServers         []string
	lokiServer          string
	loglevel            string
	egressSubjectEcs    string
	ackTimeoutS         int
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

func (c Config) IngresNatsTest() string {
	return c.ingresSubjectTest
}
func (c Config) NatsServers() string {
	return strings.Join(c.natsServers, ",")
}

func (c Config) LokiServer() string {
	return c.lokiServer
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

func (r *ConfigBuilder) withLokiServer(server *string) *ConfigBuilder {
	r.cfg.lokiServer = *server
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
func (r *ConfigBuilder) withIngresSubjectTest(ingresSubjectTest *string) *ConfigBuilder {
	r.cfg.ingresSubjectTest = *ingresSubjectTest
	return r
}

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
		errors.New("no config parsed. Use the builder at first")
	}

	return instance, nil
}

//endregion
