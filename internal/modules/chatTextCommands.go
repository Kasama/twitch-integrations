package modules

import (
	"fmt"
	"strings"
	"time"

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
	case "!estrelas":
		fallthrough
	case "!camp":
		response = "Estrelas Nascentes 3ª Edição! Venha participar do campeonato para jogadores com rank abaixo de diamante, com coaches de rank alto! Premiações em dinheiro para coaches e jogadores! https://www.start.gg/estrelasnascentes"
	// case "!matcherino":
	// 	response = "The Highest 3ª Edição! Contribua com a premimação fazendo o matcherino! https://matcherino.com/t/omegath3"
	case "!7tv":
		response = "https://7tv.app/"
	case "!desgoza", "!desgozar":
		response = msg.User.Name + " desgozou a calça do chat"
	case "!besito":
		response = "Uno besito para ti! 😘"
	case "!clt":
		response = "Prayge desemprego nosso que estáis no Brasil impessa que minha CLT seja assinada, que as ofertas de empregos caiam por terra, amém Prayge"
	case "!pig":
		t := time.Since(time.Unix(1763154500, 0))
		humanTime := int(t.Round(time.Hour * 24).Hours() / 24)
		response = fmt.Sprintf("Pig está à %d dias sem cometer atos libidinosos", humanTime)
	case "!discord":
		response = "https://discord.gg/KAJCTxzK7E"
	}

	if response != "" {
		events.Dispatch(NewChatMessage(response))
	}

	return nil
}
