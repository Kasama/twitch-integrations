package modules

import (
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
)

type CommandsModule struct {
	channel          string
	twitchChatClient *twitch.Client
}

func NewCommandsModule(channel string) *CommandsModule {
	return &CommandsModule{
		channel: channel,
	}
}

func (m *CommandsModule) Register() {
	events.Register(m.handleTwitchClient)
	events.Register(m.handleCommand)
}

func (m *CommandsModule) handleTwitchClient(client *twitch.Client) error {
	m.twitchChatClient = client
	return nil
}

func (m *CommandsModule) handleCommand(msg *twitch.PrivateMessage) error {
	if m.twitchChatClient == nil || !strings.HasPrefix(msg.Message, "!") {
		return nil
	}

	response := ""

	switch msg.Message {
	case "!7tv":
		response = "https://7tv.app/"
	}

	if response != "" {
		m.twitchChatClient.Say(m.channel, response)
	}

	return nil
}
