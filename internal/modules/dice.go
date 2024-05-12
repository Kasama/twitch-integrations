package modules

import (
	"log"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
)

type DiceModule struct {
	rolls int
}

// Register implements events.EventHandler.
func (m *DiceModule) Register() {
	events.Register(m.handleRoll)
}

func NewDiceModule() *DiceModule {
	return &DiceModule{
		rolls: 0,
	}
}

var _ events.EventHandler = &DiceModule{}

func (dm *DiceModule) handleRoll(message *twitch.PrivateMessage) error {
	if message.Message != "!roll" {
		return nil
	}
	log.Default().Printf("Rolling dice for the %d time", dm.rolls)
	dm.rolls = dm.rolls + 1
	return nil
}
