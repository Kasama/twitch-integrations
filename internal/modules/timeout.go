package modules

import (
	"fmt"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	twitchIRC "github.com/gempir/go-twitch-irc/v4"
	twitchEventSub "github.com/joeyak/go-twitch-eventsub/v2"
	helix "github.com/nicklaw5/helix/v2"
)

const rewardIDTimeout = "02346565-cfe5-41a1-821a-0f6f558e6bac"
const rewardIdTimeoutMalboro = "b4d01731-38aa-4433-af9e-cc95a3523b91"
const rewardIdForceTimeoutMalboro = "22702c49-f9b3-4f2c-b7a3-649aa8540043"
const rewardIdUnTimeoutMalboro = "c92a4c7e-6eca-4566-b4d9-254cd383a986"
const malboroID = "821524016"

const timeoutMethod = "ban"

type TimeoutModule struct {
	broadcasterID string
	client        *helix.Client
	malboroShield time.Time
	timeoutUntil  map[string]time.Time
}

func NewTimeoutModule(broadcasterID string) *TimeoutModule {
	return &TimeoutModule{
		broadcasterID: broadcasterID,
		client:        &helix.Client{},
		malboroShield: time.Unix(0, 0),
		timeoutUntil:  map[string]time.Time{},
	}
}

// timeoutUser times out the user with id user for duration seconds.
func (m *TimeoutModule) timeoutUser(user string, duration int, reason string) error {
	if timeoutMethod == "ban" {
		_, err := m.client.BanUser(&helix.BanUserParams{
			BroadcasterID: m.broadcasterID,
			ModeratorId:   m.broadcasterID,
			Body: helix.BanUserRequestBody{
				Duration: duration,
				Reason:   reason,
				UserId:   user,
			},
		})
		if err != nil {
			return err
		}
	} else {
		m.timeoutUntil[user] = time.Now().Add(time.Duration(duration) * time.Second)
	}
	return nil
}

func (m *TimeoutModule) Register() {
	events.Register(m.handleTimeoutReward)
	events.Register(m.handleUnTimeoutMalboroReward)
	events.Register(m.handleHelixClient)
	events.Register(m.handleVirtualTimeout)
}

func (m *TimeoutModule) handleHelixClient(client *helix.Client) error {
	m.client = client
	return nil
}

func (m *TimeoutModule) handleVirtualTimeout(msg *twitchIRC.PrivateMessage) error {
	if timeoutMethod == "ban" {
		return nil
	}

	if t, exists := m.timeoutUntil[msg.User.ID]; exists && time.Now().Before(t) {
		_, _ = m.client.DeleteChatMessage(&helix.DeleteChatMessageParams{
			BroadcasterID: m.broadcasterID,
			ModeratorID:   m.broadcasterID,
			MessageID:     msg.ID,
		})
	}

	return nil
}

func (m *TimeoutModule) handleUnTimeoutMalboroReward(reward *twitchEventSub.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	logger.Debugf("reward: %+v", reward)
	if reward.Reward.ID != rewardIdUnTimeoutMalboro {
		return nil
	}
	_, err := m.client.UnbanUser(&helix.UnbanUserParams{
		BroadcasterID: m.broadcasterID,
		ModeratorID:   m.broadcasterID,
		UserID:        malboroID,
	})
	if err != nil {
		return err
	}

	m.malboroShield = time.Now().Add(10 * time.Minute)
	return nil
}

func (m *TimeoutModule) handleTimeoutReward(reward *twitchEventSub.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDTimeout && reward.Reward.ID != rewardIdTimeoutMalboro && reward.Reward.ID != rewardIdForceTimeoutMalboro {
		return nil
	}

	if m.client == nil {
		return fmt.Errorf("Helix client not initialized")
	}

	user := reward.User.UserID
	duration := 600
	if reward.Reward.ID == rewardIdTimeoutMalboro || reward.Reward.ID == rewardIdForceTimeoutMalboro {
		if time.Now().Before(m.malboroShield) && reward.Reward.ID != rewardIdForceTimeoutMalboro {
			logger.Debugf("Malboro Shield is up: Denying reward")
			_, err := m.client.SendChatMessage(&helix.SendChatMessageParams{
				BroadcasterID: m.broadcasterID,
				SenderID:      m.broadcasterID,
				Message:       fmt.Sprintf("O escudo do Malboro está ativo por mais %s. Seu timeout foi em vão.", time.Until(m.malboroShield).Round(time.Second).String()),
			})
			if err != nil {
				return err
			}
			return nil
		}
		duration = 60
		user = malboroID
	}

	err := m.timeoutUser(user, duration, "timeout dos pontinhos")
	if err != nil {
		logger.Errorf("Error timing out user: %v", err)
	}

	_, err = m.client.UpdateChannelCustomRewardsRedemptionStatus(&helix.UpdateChannelCustomRewardsRedemptionStatusParams{
		ID:            reward.ID,
		BroadcasterID: m.broadcasterID,
		RewardID:      reward.Reward.ID,
		Status:        "FULFILLED",
	})
	if err != nil {
		logger.Errorf("Error updating reward status: %v", err)
	}

	return nil
}

var _ events.EventHandler = &TimeoutModule{}
