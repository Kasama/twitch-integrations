package modules

import (
	"os"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
)

type TempTextModule struct{}

func NewTempTextModule() *TempTextModule {
	return &TempTextModule{}
}

type ShowTempTextEvent struct {
	duration time.Duration
	text     string
	filepath string
}

func NewShowTempText(duration time.Duration, text string, filepath string) *ShowTempTextEvent {
	return &ShowTempTextEvent{
		duration: duration,
		text:     text,
		filepath: filepath,
	}
}

func (e *ShowTempTextEvent) WithTicker(interval time.Duration, f func()) time.Duration {
	return e.duration
}

func (e *ShowTempTextEvent) Clear() error {
	return os.WriteFile(e.filepath, []byte(""), 0644)
}

// Register implements events.EventHandler.
func (m *TempTextModule) Register() {
	events.Register(m.handleShowTempText)
}

func (m *TempTextModule) handleShowTempText(e *ShowTempTextEvent) error {
	return nil
}

var _ events.EventHandler = &TempTextModule{}
