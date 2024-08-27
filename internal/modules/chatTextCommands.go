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
		response = "Campeonato Estrelas Nascentes dia 18 de Agosto! https://start.gg/estrelasnascentes"
	case "!7tv":
		response = "https://7tv.app/"
	case "!desgoza", "!desgozar":
		response = msg.User.Name + " desgozou a calça do chat"
	case "!besito":
		response = "Uno besito para ti! 😘"
	case "!clt":
		response = "Prayge desemprego nosso que estáis no Brasil impessa que minha CLT seja assinada, que as ofertas de empregos caiam por terra, amém Prayge"
	}

	if response != "" {
		events.Dispatch(NewChatMessage(response))
	}

	return nil
}
