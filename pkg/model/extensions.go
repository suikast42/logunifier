package model

func (ecs *EcsLogEntry) HasParseErrors() bool {
	return ecs.ParseError != nil
}
