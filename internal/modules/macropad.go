package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
)

type MacropadEventState string

const (
	MacropadEventStatePressed  MacropadEventState = "pressed"
	MacropadEventStateReleased MacropadEventState = "released"
)

type MacropadEvent struct {
	Key   string
	State MacropadEventState
}

type MacropadModule struct{}

func NewMacropadModule() *MacropadModule {
	return &MacropadModule{}
}

func (m *MacropadModule) Register() {
	events.Register(m.handleMacropadEvent)
}

func (m *MacropadModule) handleMacropadEvent(event *MacropadEvent) error {
	logger.Debugf("Macropad event: %s %s", event.State, event.Key)
	return nil
}

var _ events.EventHandler = &MacropadModule{}
