package modules

import (
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
)

type CommandsModule struct {
}

func NewCommandsModule() *CommandsModule {
	return &CommandsModule{}
}

func (m *CommandsModule) Register() {
	events.Register(m.handleCommand)
}

func (m *CommandsModule) handleCommand(msg *twitch.PrivateMessage) error {
	if !strings.HasPrefix(msg.Message, "!") {
		return nil
	}

	response := ""

	switch msg.Message {
	case "!camp":
		response = "Participe do campeonato Estrelas Nascentes https://start.gg/estrelasnascentes"
	case "!7tv":
		response = "https://7tv.app/"
	case "!desgoza", "!desgozar":
		response = msg.User.Name + " desgozou a calça do chat"
	case "!besito":
		response = "Uno besito para ti! 😘"
	}

	if response != "" {
		events.Dispatch(NewChatMessage(response))
	}

	return nil
}
