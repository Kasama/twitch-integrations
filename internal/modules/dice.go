package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
)

type DiceModule struct {
	rolls int
}

func NewDiceModule() *DiceModule {
	return &DiceModule{
		rolls: 0,
	}
}

var _ events.EventHandler = &DiceModule{}

func (dm *DiceModule) ShouldHandle(_ *events.EventContext, event *events.Event) bool {
	return event.Kind == events.EventKindChatMessage && event.ChatMessage.Message == "!roll"
}

func (dm *DiceModule) HandleEvent(ctx *events.EventContext, event *events.Event) error {
	ctx.Logger.Printf("Rolling dice for the %d time", dm.rolls)
	dm.rolls = dm.rolls + 1
	return nil
}
