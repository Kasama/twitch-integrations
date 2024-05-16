package modules

import (
	"fmt"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/gempir/go-twitch-irc/v4"
)

type YappingModule struct {
	yapCount     map[string]int
	channel      string
	twitchClient *twitch.Client
}

// Register implements events.EventHandler.
func (m *YappingModule) Register() {
	events.Register(m.handleTwitchClient)
	events.Register(m.handlePrivateMessage)
}

func NewYappingModule(channel string) *YappingModule {
	return &YappingModule{
		yapCount:     make(map[string]int, 0),
		channel:      channel,
		twitchClient: nil,
	}
}

func treatUserName(user string) string {
	return strings.TrimPrefix(strings.TrimSpace(strings.ToLower(user)), "@")
}

func (m *YappingModule) handleTwitchClient(client *twitch.Client) error {
	logger.Debug("YappingModule: Twitch client initialized with client")
	m.twitchClient = client
	return nil
}

func (m *YappingModule) handlePrivateMessage(message *twitch.PrivateMessage) error {
	if m.twitchClient == nil {
		return fmt.Errorf("twitch client not initialized, but got message event")
	}
	author := treatUserName(message.User.DisplayName)
	m.yapCount[author] = m.yapCount[author] + 1

	fields := strings.Fields(message.Message)
	if fields[0] == "!fala" {
		if len(fields) < 2 {
			m.twitchClient.Say(m.channel, "Uso: !fala @<usuário>. Mostra quantas vezes o usuário falou no chat hoje")
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

		m.twitchClient.Say(m.channel, message)
	}
	return nil
}

var _ events.EventHandler = &YappingModule{}
