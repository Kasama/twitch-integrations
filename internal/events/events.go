package events

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/badu/bus"
)

func Dispatch[T any](event T) {
	logger.Debugf("Dispatching event '%T'", event)
	bus.Pub(event)
}

func Register[T any](callback func(event T) error) {
	bus.Sub(func(event T) {
		err := callback(event)
		if err != nil {
			logger.Errorf("Error processing event '%T': %s\n", event, err.Error())
		}
	})
}

type EventHandler interface {
	Register()
}

type EventEmitter interface {
	// Start starts the event emitter in a new goroutine. Closing the returned channel will stop the event emitter
	Start() chan struct{}
	// Stop stops the event emitter
	Stop(chan struct{})
}
