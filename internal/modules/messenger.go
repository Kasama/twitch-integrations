package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
)

type Message struct {
	msg string
}

func NewChatMessage(msg string) *Message {
	return &Message{
		msg: msg,
	}
}

type MessengerModule struct {
	channelID    string
	twitchClient *twitch.Client
}

func NewMessengerModule(channelID string) *MessengerModule {
	return &MessengerModule{
		channelID:    channelID,
		twitchClient: nil,
	}
}

func (m *MessengerModule) Register() {
	events.Register(m.handleTwitchClient)
	events.Register(m.handleMessage)
}

func (m *MessengerModule) handleTwitchClient(client *twitch.Client) error {
	m.twitchClient = client
	return nil
}

func (m *MessengerModule) handleMessage(msg *Message) error {
	if m.twitchClient == nil {
		return nil
	}
	m.twitchClient.Say(m.channelID, msg.msg)
	return nil
}
