package events

import "log"

type EventOptions func(*EventContext)

func WithLogger(logger *log.Logger) EventOptions {
	return func(ctx *EventContext) {
		ctx.Logger = logger
	}
}
