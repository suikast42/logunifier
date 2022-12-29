package patterns

import (
	"errors"
	"fmt"
	"github.com/trivago/grok"
	additionalPatterns "github.com/trivago/grok/patterns"
	"strings"
	"sync"
)

type ParseResult struct {
	LogLevel  string
	TimeStamp string
	Msg       string
}

func (p ParseResult) String() string {
	return fmt.Sprintf("LogLevel: %s Timestamp: %s Message: %s", p.TimeStamp, p.LogLevel, p.Msg)
}

type PatternKey string
type atributeKeys string

const (
	timestamp atributeKeys = "timestamp"
	message   atributeKeys = "message"
	level     atributeKeys = "level"
)
const (
	TS_LEVEL_MSG PatternKey = "TS_LEVEL_MSG"
	MSG_ONLY     PatternKey = "MSG"
)

var APPLOGS = map[string]string{
	"MULTILINE":          `((\s)*(.*))*`,
	string(MSG_ONLY):     `%{MULTILINE:message}`,
	string(TS_LEVEL_MSG): `%{TIMESTAMP_ISO8601:timestamp} \[%{LOGLEVEL:level}\] %{MULTILINE:message}`,
}

type PatternFactory struct {
	patterns  map[string]string
	compilers map[string]*grok.CompiledGrok
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

	// end::tagname[]
	addPatterns := make(map[string]string)
	compiledPatterns := make(map[string]*grok.CompiledGrok)
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
		err := add(addPatterns, APPLOGS)
		if err != nil {
			panic(err)
		}
	}
	//{
	//	err := add(addPatterns, additionalPatterns.Java)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.LinuxSyslog)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.PostgreSQL)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.Rails)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.Redis)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, additionalPatterns.Ruby)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//{
	//	err := add(addPatterns, NGNIX)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	config := grok.Config{
		Patterns:            addPatterns,
		SkipDefaultPatterns: true,
	}
	g, err := grok.New(config)
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

	instance = &PatternFactory{
		patterns:  addPatterns,
		compilers: compiledPatterns,
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

func (factory *PatternFactory) Parse(patternkey PatternKey, text string) (ParseResult, error) {
	compiledGrok := factory.compilers[string(patternkey)]
	if compiledGrok == nil {
		return ParseResult{}, errors.New(fmt.Sprintf("No compiler found for key %s", patternkey))
	}
	parsed := compiledGrok.ParseString(text)

	return ParseResult{
		LogLevel:  parsed[string(level)],
		TimeStamp: parsed[string(timestamp)],
		Msg:       parsed[string(message)],
	}, nil

}
