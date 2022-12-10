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
	natsServers         []string
	loglevel            string
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

func (c Config) NatsServers() []string {
	return c.natsServers
}

func (c Config) Loglevel() string {
	return c.loglevel
}

//endregion

// region enums
type IngressSubject string

const (
	IngressLogsJournald IngressSubject = "ingress.logs.journald"
)

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

func (r *ConfigBuilder) withLogLevel(loglevel *string) *ConfigBuilder {
	r.cfg.loglevel = *loglevel
	return r
}

func (r *ConfigBuilder) withIngress(ingress *string, _type IngressSubject) *ConfigBuilder {
	switch _type {
	case IngressLogsJournald:
		r.cfg.ingressNatsJournald = *ingress
	}
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
