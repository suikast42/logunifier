package patterns

import (
	"fmt"
	"github.com/suikast42/logunifier/pkg/model"
	"github.com/suikast42/logunifier/pkg/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"time"
)

type GrokPatternLogfmt struct {
	GrokPatternDefault
	// Builder fields
	_parseErrors []string
	_logfmtKv    map[string]string
}

var logFmtTimeStanpkeys = []string{"ts,timestamp"}

func (g *GrokPatternLogfmt) from(log *model.MetaLog) GrokPatternExtractor {
	g._metaLog = log
	g._this = g
	g._logfmtKv = map[string]string{}
	logMessage, err := utils.DecodeLogFmt(log.Message)
	if err != nil {
		g._parseErrors = append(g._parseErrors, err.Error())
		return g._this
	}
	for k, v := range logMessage {
		val, ok := g._logfmtKv[k]
		// If the key exists
		if ok {
			g._parseErrors = append(g._parseErrors, fmt.Sprintf("The key %s already exists. Override the value %s", k, val))
		}
		g._logfmtKv[k] = v
	}
	return g._this
}

func (g *GrokPatternLogfmt) timeStamp() GrokPatternExtractor {
	tsstring, ok := g._logfmtKv[string(utils.LogfmtKeyTimestamp)]
	var parsedTs time.Time
	if !ok {
		g._parseErrors = append(g._parseErrors, "Can't find timestamp key in msg")
		return g._this
	}
	defer func() {
		delete(g._logfmtKv, string(utils.LogfmtKeyTimestamp))
	}()
	cachedLayout, found := cachedLayoutForLog(g._metaLog)
	if !found {
		for _, layout := range g.TimeStampFormats {
			_parsed, err := time.Parse(layout, tsstring)
			if err != nil {
				continue
			}
			parsedTs = _parsed
			cacheLayoutForLog(g._metaLog, layout)
		}
	} else {
		_parsed, err := time.Parse(cachedLayout, tsstring)
		if err != nil {
			// Delete the ts from chance and retry all again
			deleteCachedLayoutForLog(g._metaLog)
			return g.timeStamp()
		}
		parsedTs = _parsed
	}

	if parsedTs.IsZero() {
		g._parseErrors = append(g._parseErrors, fmt.Sprintf("Can't find timestamp for %s", tsstring))
		return g._this
	}
	g._timeStamp = timestamppb.New(parsedTs)

	return g._this
}

func (g *GrokPatternLogfmt) message() GrokPatternExtractor {

	message, ok := g._logfmtKv[string(utils.LogfmtKeyMessage)]
	if !ok {
		g._parseErrors = append(g._parseErrors, "Can't find a message key in msg")
		return g._this
	}
	defer func() {
		delete(g._logfmtKv, string(utils.LogfmtKeyMessage))
	}()
	g._message = message
	return g._this
}

func (g *GrokPatternLogfmt) errorInfo() GrokPatternExtractor {
	errMsg, ok := g._logfmtKv[string(utils.LogfmtKeyError)]
	if ok {
		defer func() {
			delete(g._logfmtKv, string(utils.LogfmtKeyError))
		}()
		err := model.Error{
			Code:       "",
			Id:         "",
			Message:    errMsg,
			StackTrace: "",
			Type:       "",
		}
		g._errorInfo = &err
	}

	return g._this
}

func (g *GrokPatternLogfmt) logInfo() GrokPatternExtractor {
	caller, callerFound := g._logfmtKv[string(utils.LogfmtKeyCaller)]
	g._logInfo = &model.Log{
		File:       nil,
		Level:      g._metaLog.FallbackLoglevel,
		Logger:     "",
		ThreadName: "",
		Original:   "",
		Syslog:     nil,
		LevelEmoji: model.LogLevelToEmoji(g._metaLog.FallbackLoglevel),
	}
	if callerFound {
		defer func() {
			delete(g._logfmtKv, string(utils.LogfmtKeyCaller))
		}()
		split := strings.Split(caller, ":")
		var log = split[0]
		var line = int32(-1)
		if len(split) == 2 {
			_line, err := strconv.ParseInt(split[1], 10, 32)
			if err != nil {
				line = int32(_line)
			}
		}
		g._logInfo.Origin = &model.Log_Origin{
			File: &model.Log_Origin_File{
				Line: line,
				Name: log,
			},
			Function: "",
		}
	}

	level, levelFound := g._logfmtKv[string(utils.LogfmtKeyLevel)]
	if levelFound {
		defer func() {
			delete(g._logfmtKv, string(utils.LogfmtKeyLevel))
		}()
		g._logInfo.Level = model.StringToLogLevel(level)
		g._logInfo.LevelEmoji = model.LogLevelToEmoji(g._logInfo.Level)
	}
	return g._this

}
func (g *GrokPatternLogfmt) userInfo() GrokPatternExtractor {
	user, ok := g._logfmtKv[string(utils.LogfmtKeyUser)]
	if ok {
		defer func() {
			delete(g._logfmtKv, string(utils.LogfmtKeyUser))
		}()
		g._userInfo = &model.User{Name: user}
	}

	return g._this
}

func (g *GrokPatternLogfmt) eventInfo() GrokPatternExtractor {
	kind, ok := g._logfmtKv[string(utils.LogfmtKeyEvent)]
	if ok {
		defer func() {
			delete(g._logfmtKv, string(utils.LogfmtKeyEvent))
		}()
		g._eventInfo = &model.Event{Kind: kind}
	}
	return g._this
}

func (g *GrokPatternLogfmt) tracingInfo() GrokPatternExtractor {
	traceid, ok := g._logfmtKv[string(utils.LogfmtKeyTraceID)]
	spanid, _ := g._logfmtKv[string(utils.LogfmtKeySpanID)]
	if ok {
		defer func() {
			delete(g._logfmtKv, string(utils.LogfmtKeyTraceID))
			delete(g._logfmtKv, string(utils.LogfmtKeySpanID))
		}()
		g._traceInfo = &model.Tracing{
			Span:        &model.Tracing_Span{Id: spanid},
			Trace:       &model.Tracing_Trace{Id: traceid},
			Transaction: nil,
		}
	}
	return g._this
}

func (g *GrokPatternLogfmt) extract() *model.EcsLogEntry {
	ecs := model.NewEcsLogEntry()
	ecs.Timestamp = g._timeStamp
	ecs.Log = g._logInfo
	ecs.Message = g._message
	ecs.Labels = g._labels
	ecs.Tags = g._tags
	ecs.Container = g._containerInfo
	ecs.Agent = g._agentInfo
	ecs.Host = g._hostInfo
	ecs.Organization = g._organisationInfo
	ecs.Service = g._serviceInfo
	ecs.Error = g._errorInfo
	ecs.Trace = g._traceInfo
	ecs.ProcessError = g._metaLog.ProcessError

	// Every step removes the registered keys
	// Add the not standard keys as labels
	for k, v := range g._logfmtKv {
		ecs.Labels["logfmt_"+k] = v
	}

	return ecs
}
