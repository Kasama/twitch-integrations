package modules

import (
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
)

type ApostaModule struct {
	broadcasterID string
	helix         *helix.Client
}

func NewApostaModule(broadcasterID string) *ApostaModule {
	return &ApostaModule{
		broadcasterID: broadcasterID,
	}
}

func (m *ApostaModule) Register() {
	events.Register(m.handleMessage)
	events.Register(m.handleHelix)
}

func (m *ApostaModule) handleHelix(client *helix.Client) error {
	m.helix = client
	return nil
}

func (m *ApostaModule) handleMessage(msg *twitch.PrivateMessage) error {
	if !strings.HasPrefix(msg.Message, "!aposta") || m.helix == nil {
		return nil
	}

	content := strings.TrimPrefix(msg.Message, "!aposta ")

	parts := strings.Split(content, ",")

	question := parts[0]

	if strings.TrimSpace(question) == "" {
		return nil
	}

	option1 := "sim"
	option2 := "não"
	if parts == nil || len(parts) >= 3 {
		option1 = parts[1]
		option2 = parts[2]
	}

	_, err := m.helix.CreatePrediction(&helix.CreatePredictionParams{
		BroadcasterID: m.broadcasterID,
		Title:         question,
		Outcomes: []helix.PredictionChoiceParam{
			{
				Title: option1,
			},
			{
				Title: option2,
			},
		},
		PredictionWindow: 0,
	})

	return err
}
