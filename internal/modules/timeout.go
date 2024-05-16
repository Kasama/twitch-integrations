package modules

import (
	"fmt"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	twitchEventSub "github.com/joeyak/go-twitch-eventsub/v2"
	helix "github.com/nicklaw5/helix/v2"
)

const rewardIDTimeout = "02346565-cfe5-41a1-821a-0f6f558e6bac"

type TimeoutModule struct {
	broadcasterID string
	client        *helix.Client
}

func NewTimeoutModule(broadcasterID string) *TimeoutModule {
	return &TimeoutModule{
		broadcasterID: broadcasterID,
	}
}

func (m *TimeoutModule) Register() {
	events.Register(m.handleReward)
	events.Register(m.handleHelixClient)
}

func (m *TimeoutModule) handleHelixClient(client *helix.Client) error {
	m.client = client
	return nil
}

func (m *TimeoutModule) handleReward(reward *twitchEventSub.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDTimeout {
		return nil
	}

	if m.client == nil {
		return fmt.Errorf("Helix client not initialized")
	}

	resp, err := m.client.BanUser(&helix.BanUserParams{
		BroadcasterID: m.broadcasterID,
		ModeratorId:   m.broadcasterID,
		Body: helix.BanUserRequestBody{
			Duration: 600,
			Reason:   "pediu timeout",
			UserId:   reward.User.UserID,
		},
	})
	if err != nil {
		return err
	}
	logger.Debugf("Banned Users: %+v", resp)

	return nil
}

var _ events.EventHandler = &TimeoutModule{}
