package httpclient

import (
	"github.com/rs/zerolog"
)

type retryableHTTPLeveledLoggerAgapter struct {
	logger *zerolog.Logger
}

func (a *retryableHTTPLeveledLoggerAgapter) Debug(msg string, keysAndValues ...interface{}) {
	event := a.logger.Debug()
	event = addKeysAndValues(event, keysAndValues...)
	event.Msg(msg)
}

func (a *retryableHTTPLeveledLoggerAgapter) Error(msg string, keysAndValues ...interface{}) {
	event := a.logger.Error()
	event = addKeysAndValues(event, keysAndValues...)
	event.Msg(msg)
}

func (a *retryableHTTPLeveledLoggerAgapter) Info(msg string, keysAndValues ...interface{}) {
	event := a.logger.Info()
	event = addKeysAndValues(event, keysAndValues...)
	event.Msg(msg)
}

func (a *retryableHTTPLeveledLoggerAgapter) Warn(msg string, keysAndValues ...interface{}) {
	event := a.logger.Warn()
	event = addKeysAndValues(event, keysAndValues...)
	event.Msg(msg)
}

func addKeysAndValues(event *zerolog.Event, keysAndValues ...interface{}) *zerolog.Event {
	for i := 0; i < len(keysAndValues); i += 2 {
		if key, ok := keysAndValues[i].(string); ok {
			event = event.Interface(key, keysAndValues[i+1])
		}
	}
	return event
}
