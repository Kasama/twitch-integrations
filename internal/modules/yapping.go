package modules

import (
	"fmt"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
)

type YappingModule struct {
	yapCount     map[string]int
	channel      string
	twitchClient *twitch.Client
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

func (m *YappingModule) HandleEvent(ctx *events.EventContext, event *events.Event) error {
	switch event.Kind {
	case events.EventKindChatClientCreated:
		{
			m.twitchClient = event.ChatClientCreated
		}
	case events.EventKindChatMessage:
		{
			if m.twitchClient == nil {
				return fmt.Errorf("twitch client not initialized, but got message event")
			}
			message := event.ChatMessage
			author := treatUserName(message.User.DisplayName)
			m.yapCount[author] = m.yapCount[author] + 1

			fields := strings.Fields(message.Message)
			if fields[0] == "!fala" {
				if len(fields) < 2 {
					m.twitchClient.Say(m.channel, "Uso: !fala <usuário>")
					m.twitchClient.Say(m.channel, "Mostra quantas vezes o usuário falou no chat hoje")
					return nil
				}
				user := treatUserName(fields[1])
				count := m.yapCount[user]
				m.twitchClient.Say(m.channel, fmt.Sprintf("%s has yapped %d times", user, count))
			}
		}
	}

	return nil
}

func (*YappingModule) ShouldHandle(ctx *events.EventContext, event *events.Event) bool {
	return event.Kind == events.EventKindChatMessage || event.Kind == events.EventKindChatClientCreated
}

var _ events.EventHandler = &YappingModule{}
