package events

import (
	"log"

	twitchChat "github.com/gempir/go-twitch-irc/v4"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

const (
	EventKindChatMessage                   = "chat_message"
	EventKindChatClientCreated             = "chat_client_created"
	EventKindStop                          = "stop"
	EventKindConnected                     = "connected"
	EventKindChannelPointsRewardRedemption = "channel_points_reward_redemption"
)

type Event struct {
	Kind                          string
	ChatMessage                   *twitchChat.PrivateMessage
	ChatClientCreated             *twitchChat.Client
	ChannelPointsRewardRedemption *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd
}

type EventContext struct {
	Logger          *log.Logger
	EventDispatcher *EventDispatcher
}

func newEventContext() *EventContext {
	return &EventContext{
		Logger: log.Default(),
	}
}

type EventHandler interface {
	ShouldHandle(ctx *EventContext, event *Event) bool
	HandleEvent(ctx *EventContext, event *Event) error
}

// TempEventHandler is a handler that is removed after one use. It should return true if it is done, if not done, it will be scheduled to be called again
type TempEventHandler = func(ctx *EventContext, event *Event) bool

type EventEmitter interface {
	// Start starts the event emitter in a new goroutine. Closing the returned channel will stop the event emitter
	Start(dispatcher *EventDispatcher) chan struct{}
	// Stop stops the event emitter
	Stop(chan struct{})
}

type EventDispatcher struct {
	eventQueue   chan *Event
	handlers     []EventHandler
	tempHandlers []TempEventHandler
	Context      *EventContext
}

func NewEventDispatcher(opts ...EventOptions) *EventDispatcher {
	ctx := newEventContext()
	for _, opt := range opts {
		opt(ctx)
	}

	dispatcher := &EventDispatcher{
		eventQueue:   make(chan *Event),
		handlers:     make([]EventHandler, 0),
		tempHandlers: make([]TempEventHandler, 0),
		Context:      ctx,
	}

	dispatcher.Context.EventDispatcher = dispatcher

	return dispatcher
}

func (ed *EventDispatcher) Dispatch(event *Event) {
	ed.eventQueue <- event
}

func (ed *EventDispatcher) RegisterHandler(handler EventHandler) {
	ed.handlers = append(ed.handlers, handler)
}

func (ed *EventDispatcher) RegisterTempHandler(handler TempEventHandler) {
	ed.tempHandlers = append(ed.tempHandlers, handler)
}

func (ed *EventDispatcher) Start() {
	for event := range ed.eventQueue {
		if event.Kind == EventKindStop {
			break
		}
		newTempHandlers := make([]TempEventHandler, 0)
		for _, tempHandler := range ed.tempHandlers {
			done := tempHandler(ed.Context, event)
			if !done {
				newTempHandlers = append(newTempHandlers, tempHandler)
			}
		}
		ed.tempHandlers = newTempHandlers
		for _, handler := range ed.handlers {
			if handler.ShouldHandle(ed.Context, event) {
				err := handler.HandleEvent(ed.Context, event)
				if err != nil {
					log.Printf("error handling event: %v\n", err)
				}
			}
		}
	}
}

func (ed *EventDispatcher) StartAsync() chan struct{} {
	done := make(chan struct{})
	go func() {
		ed.Start()
		close(done)
	}()
	return done
}

func (ed *EventDispatcher) Stop() {
	ed.eventQueue <- &Event{
		Kind: EventKindStop,
	}
}
