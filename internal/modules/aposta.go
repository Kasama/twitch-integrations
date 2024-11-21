package modules

import (
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
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

	content := strings.TrimPrefix(msg.Message, "!aposta")

	parts := strings.Split(content, ",")

	question := strings.TrimSpace(parts[0])

	if question == "" {
		events.Dispatch(NewChatMessage("use !aposta <opção 1>, <opção 2>, <pergunta>"))
		return nil
	}

	option1 := "sim"
	option2 := "não"
	if parts == nil || len(parts) >= 3 {
		option1 = strings.TrimSpace(parts[1])
		option2 = strings.TrimSpace(parts[2])
	}

	predictionResponse, err := m.helix.CreatePrediction(&helix.CreatePredictionParams{
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
		PredictionWindow: int((10 * time.Minute).Seconds()),
	})

	if predictionResponse.ErrorMessage != "" {
		logger.Errorf("Problem with prediction: '%s'", predictionResponse.ErrorMessage)
	}

	return err
}
