package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	helix "github.com/nicklaw5/helix/v2"
)

type TwitchHelixModule struct {
	client *helix.Client
}

func NewTwitchHelixModule() *TwitchHelixModule {
	return &TwitchHelixModule{}
}

// Register implements events.EventHandler.
func (m *TwitchHelixModule) Register() {
	events.Register(m.handleAuth)
}

func (m *TwitchHelixModule) handleAuth(auth *twitch.TwitchAuth) error {
	client, err := helix.NewClient(&helix.Options{
		ClientID:        auth.TwitchConfig.ClientId,
		ClientSecret:    auth.TwitchConfig.ClientSecret,
		UserAccessToken: auth.AccessToken,
		UserAgent:       "kasama-twitch-integrations",
	})

	if err != nil {
	  return err
	}

	m.client = client
	events.Dispatch(m.client)

	return nil
}

var _ events.EventHandler = &TwitchHelixModule{}
