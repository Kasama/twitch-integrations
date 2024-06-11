package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
)

type BesitoModule struct {
	twitchChatClient *twitch.Client
}

func (m *BesitoModule) Register() {
	events.Register(m.handleClient)
	events.Register(m.handleBesito)
}

func NewBesitoModule() *BesitoModule {
	return &BesitoModule{}
}

func (m *BesitoModule) handleClient(client *twitch.Client) error {
	m.twitchChatClient = client
	return nil
}

func (m *BesitoModule) handleBesito(msg *twitch.PrivateMessage) error {
	if msg.Message != "!besito" || m.twitchChatClient == nil {
		return nil
	}

	m.twitchChatClient.Say(msg.Channel, "Uno besito para ti! 😘")
	return nil
}
