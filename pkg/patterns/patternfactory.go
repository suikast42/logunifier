package patterns

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/suikast42/logunifier/internal/config"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/utils"
	"github.com/trivago/grok"
	additionalPatterns "github.com/trivago/grok/patterns"
	"strings"
	"sync"
)

type PatternFactory struct {
	logger    *zerolog.Logger
	patterns  map[string]string
	compilers map[string]*grok.CompiledGrok
}

func (factory *PatternFactory) CompilerFor(key model.MetaLog_PatternKey) *grok.CompiledGrok {
	return factory.compilerFor(key.String())
}

func (factory *PatternFactory) compilerFor(key string) *grok.CompiledGrok {
	return factory.compilers[key]
}

var mtx sync.Mutex

var instance *PatternFactory

func Instance() *PatternFactory {
	return instance
}
func Initialize() (*PatternFactory, error) {
	mtx.Lock()
	defer mtx.Unlock()
	if instance != nil {
		return instance, nil
	}
	compiledPatterns := make(map[string]*grok.CompiledGrok)
	addPatterns := make(map[string]string)

	{
		err := add(addPatterns, grok.DefaultPatterns)
		// better defined in additionalPatterns.Grok
		for k := range additionalPatterns.Grok {
			val := grok.DefaultPatterns[k]
			if len(val) > 0 {
				delete(addPatterns, k)
			}
		}
		if err != nil {
			panic(err)
		}
	}
	{
		err := add(addPatterns, additionalPatterns.Grok)
		if err != nil {
			panic(err)
		}
	}

	{
		err := add(addPatterns, utils.CustomPatterns)
		if err != nil {
			panic(err)
		}
	}
	grokConfig := grok.Config{
		Patterns:            addPatterns,
		SkipDefaultPatterns: true,
	}
	g, err := grok.New(grokConfig)
	if err != nil {
		suberror := errors.New(fmt.Sprintf("Can't create new isnatce\n%s", err.Error()))
		return nil, suberror
	}
	for k, v := range addPatterns {
		compiled, err := g.Compile(v)
		if err != nil {
			suberror := errors.New(fmt.Sprintf("Cannot compile %s with value %s\n%s", k, v, err.Error()))
			return nil, suberror
		}
		compiledPatterns[k] = compiled
	}
	logger := config.Logger()
	instance = &PatternFactory{
		patterns:  addPatterns,
		compilers: compiledPatterns,
		logger:    &logger,
	}

	return instance, nil
}

func add(source map[string]string, new map[string]string) error {

	for k, v := range new {
		//key := prefix + "_" + k
		key := k
		// IF the key is already present and the default is different
		if val, ok := source[key]; ok && !strings.EqualFold(val, v) {
			return errors.New(fmt.Sprintf("%s already exists for value %s and %s should aded", key, v, val))
		}
		source[key] = v
	}
	return nil
}

func (factory *PatternFactory) Parse(log *model.MetaLog) *model.EcsLogEntry {
	if log.HasProcessErrors() {
		log.EcsLogEntry.Message = "Can't parse a MetaLog with process errors. See the ProcessError Raw message for further debugging"
		log.EcsLogEntry.SetLogLevel(model.LogLevel_fatal)
		return log.EcsLogEntry
	}
	// Native Ecs.
	if log.PatternKey == model.MetaLog_Ecs {
		return log.EcsLogEntry
	}
	extractor := factory.findPatternFor(log)
	return ExtractFrom(extractor, log)
}

func (factory *PatternFactory) findPatternFor(log *model.MetaLog) GrokPatternExtractor {

	switch log.PatternKey {
	case model.MetaLog_LogFmt:
		return &GrokPatternLogfmt{
			GrokPatternDefault: GrokPatternDefault{
				GrokPattern: GrokPattern{
					Name: log.PatternKey,
				},
			},
		}
	case model.MetaLog_TsLevelMsg,
		model.MetaLog_Traefik,
		model.MetaLog_Envoy:
		return &GrokPatternTsLevelMsg{
			GrokPatternDefault: GrokPatternDefault{
				GrokPattern: GrokPattern{
					Name: log.PatternKey,
				},
			},
		}

		//case model.MetaLog_Ecs:
	case model.MetaLog_Nop:
		return &GrokPatternDefault{
			GrokPattern: GrokPattern{
				Name: model.MetaLog_Nop,
			},
		}
	default:
		log.AppendParseError(fmt.Sprintf("The identified PatternKey %s by the ingress is not mapped to a pattern extractor", log.PatternKey.String()))
		return &GrokPatternDefault{
			GrokPattern: GrokPattern{
				Name: model.MetaLog_Nop,
			},
		}
	}

}
