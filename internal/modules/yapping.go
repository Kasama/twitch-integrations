package modules

import (
	"fmt"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/gempir/go-twitch-irc/v4"
)

type YappingModule struct {
	yapCount map[string]int
}

// Register implements events.EventHandler.
func (m *YappingModule) Register() {
	events.Register(m.handleTwitchClient)
	events.Register(m.handlePrivateMessage)
}

func NewYappingModule() *YappingModule {
	return &YappingModule{
		yapCount: make(map[string]int, 0),
	}
}

func treatUserName(user string) string {
	return strings.TrimPrefix(strings.TrimSpace(strings.ToLower(user)), "@")
}

func (m *YappingModule) handleTwitchClient(client *twitch.Client) error {
	logger.Debug("YappingModule: Twitch client initialized with client")
	return nil
}

func (m *YappingModule) handlePrivateMessage(message *twitch.PrivateMessage) error {
	author := treatUserName(message.User.DisplayName)
	m.yapCount[author] = m.yapCount[author] + 1

	fields := strings.Fields(message.Message)
	if fields[0] == "!fala" {
		if len(fields) < 2 {
			events.Dispatch(NewChatMessage("Uso: !fala @<usuário>. Mostra quantas vezes o usuário falou no chat hoje"))
			return nil
		}
		user := treatUserName(fields[1])
		count := m.yapCount[user]

		message := ""
		if count == 0 {
			message = fmt.Sprintf("%s não falou nada hoje", user)
		} else if count < 10 {
			message = fmt.Sprintf("%s falou só %d coisas hoje", user, count)
		} else {
			message = fmt.Sprintf("%s não para quieto, já foram %d bostas hoje itskas19Yapping", user, count)
		}

		events.Dispatch(NewChatMessage(message))
	}
	return nil
}

var _ events.EventHandler = &YappingModule{}
