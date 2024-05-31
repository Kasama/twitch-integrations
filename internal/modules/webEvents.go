package modules

import (
	"fmt"
	"io"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type WebEvent struct {
	name string
	data string
}

func NewWebEvent(name, data string) *WebEvent {
	return &WebEvent{
		name: name,
		data: data,
	}
}

func (e *WebEvent) Dispatch() {
	events.Dispatch(e)
}

func (e *WebEvent) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.name, strings.ReplaceAll(e.data, "\n", ""))
	if err != nil {
		return err
	}

	return nil
}

type WebEventsModule struct {
	consumers map[uuid.UUID]chan *WebEvent
}

func NewWebEventsModule() *WebEventsModule {
	return &WebEventsModule{
		consumers: make(map[uuid.UUID]chan *WebEvent),
	}
}

func (m *WebEventsModule) Register() {
	events.Register(m.handleWebEvent)
}

func (m *WebEventsModule) HandleSSE(c echo.Context) error {

	resp := c.Response()
	resp.Header().Set("Content-Type", "text/event-stream")
	resp.Header().Set("Cache-Control", "no-cache")
	resp.Header().Set("Connection", "keep-alive")

	eventConsumer := make(chan *WebEvent)
	myid := uuid.New()
	m.consumers[myid] = eventConsumer

	for {
		select {
		case <-c.Request().Context().Done():
			logger.Debugf("Reached end of context for SSE connection")
			delete(m.consumers, myid)
			return nil
		case e := <-eventConsumer:
			err := e.Write(resp)
			resp.Flush()
			if err != nil {
				return err
			}
		}
	}
}

func (m *WebEventsModule) handleWebEvent(event *WebEvent) error {
	logger.Debugf("got event: %v", event)
	if event != nil {
		for _, consumer := range m.consumers {
			go func(c chan *WebEvent) {
				c <- event
			}(consumer)
		}
	}

	return nil
}

var _ events.EventHandler = &WebEventsModule{}
