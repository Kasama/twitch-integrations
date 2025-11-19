package modules

import (
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	twitchChat "github.com/gempir/go-twitch-irc/v4"
	"github.com/joeyak/go-twitch-eventsub/v2"
	helix "github.com/nicklaw5/helix/v2"
)

const rewardIDMalboroRush = "c1a3f6b8-393a-45c4-897b-afd2da608546"

type MalboroRustEvent struct {
	Duration time.Duration
}

type MalboroRushModule struct {
	helixClient   *helix.Client
	active        bool
	broadcasterID string
}

func NewMalboroRushModule(broadcasterID string) *MalboroRushModule {
	return &MalboroRushModule{
		active:        false,
		broadcasterID: broadcasterID,
	}
}

func (m *MalboroRushModule) Register() {
	events.Register(m.handleMessage)
	events.Register(m.handleReward)
	events.Register(m.handleTwitchClient)
}

func (m *MalboroRushModule) handleTwitchClient(client *helix.Client) error {
	m.helixClient = client
	return nil
}

func (m *MalboroRushModule) handleMessage(msg *twitchChat.PrivateMessage) error {
	_, isMod := msg.User.Badges["moderator"]
	_, isBroadcaster := msg.User.Badges["broadcaster"]
	if isBroadcaster || isMod || m.helixClient == nil {
		return nil
	}
	if !m.active {
		return nil
	}


	content := strings.ToLower(msg.Message)
	if strings.Contains(content, "malboro") {
		return nil
	}

	_, err := m.helixClient.BanUser(&helix.BanUserParams{
		BroadcasterID: m.broadcasterID,
		ModeratorId:   m.broadcasterID,
		Body: helix.BanUserRequestBody{
			Duration: 60,
			Reason:   "Não falou malboro durante o malboro rush",
			UserId:   msg.User.ID,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *MalboroRushModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	logger.Debugf("Got a reward %+v", reward)
	if reward.Reward.ID != rewardIDMalboroRush {
		return nil
	}

	logger.Debugf("começando malboro rush")
	m.active = true

	events.Dispatch(NewChatMessage("Malboro Rush começou! Pelo proximo minuto, todas as mensagens devem conter 'malboro' ou sofrer as consequencias"))

	go func() {
		duration := time.Minute
		timer := time.NewTimer(duration)

		<- timer.C

		logger.Debugf("acabando malboro rush")
		m.active = false
		events.Dispatch(NewChatMessage("Malboro Rush acabou!"))
	}()

	return nil
}
