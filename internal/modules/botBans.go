package modules

import (
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	twitchChat "github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
)

type BotBansModule struct {
	broadcasterID string
	helixClient   *helix.Client
}

func NewBotBansModule(broadcasterID string) *BotBansModule {
	return &BotBansModule{
		broadcasterID: broadcasterID,
		helixClient:   nil,
	}
}

func (m *BotBansModule) Register() {
	events.Register(m.handleHelix)
	events.Register(m.handleBan)
}

func (m *BotBansModule) handleHelix(client *helix.Client) error {
	m.helixClient = client
	return nil
}

func (m *BotBansModule) handleBan(msg *twitchChat.PrivateMessage) error {

	bannedWords := []string{
		"view", "viewer", "viewers", "buy",
	}

	banneable := false

	if msg.FirstMessage {
		for _, word := range bannedWords {
			if strings.Contains(strings.ToLower(msg.Message), word) {
				banneable = true
				break
			}
		}
	}

	var err error
	if banneable && m.helixClient != nil {
		_, err = m.helixClient.BanUser(&helix.BanUserParams{
			BroadcasterID: m.broadcasterID,
			ModeratorId:   m.broadcasterID,
			Body: helix.BanUserRequestBody{
				Reason: "Bot de ads",
				UserId: msg.User.ID,
			},
		})
	}

	return err
}
