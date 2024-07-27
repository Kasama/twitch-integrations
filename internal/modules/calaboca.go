package modules

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/andreykaipov/goobs"
	inputRequest "github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/joeyak/go-twitch-eventsub/v2"
	"github.com/nicklaw5/helix/v2"
)

const rewardIDMuteMe = "1bbe8c0f-d7f5-48f2-a658-7b6ad2a2546d"

type EventMuteMe struct {
	Duration time.Duration
	User     string
}

type CalabocaModule struct {
	obs           *goobs.Client
	helix         *helix.Client
	broadcasterID string
}

func NewCalabocaModule(broadcasterID string) *CalabocaModule {
	return &CalabocaModule{
		broadcasterID: broadcasterID,
	}
}

func (m *CalabocaModule) Register() {
	events.Register(m.handleOBSClient)
	events.Register(m.handleReward)
	events.Register(m.handleMuteMe)
	events.Register(m.handleTwitchHelix)
}

func (m *CalabocaModule) handleOBSClient(client *goobs.Client) error {
	m.obs = client

	return nil
}

func (m *CalabocaModule) handleTwitchHelix(client *helix.Client) error {
	m.helix = client

	return nil
}

func (m *CalabocaModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDMuteMe {
		return nil
	}
	events.Dispatch(NewPunishableRedeemInfo(reward.User.UserID, reward.User.UserName))
	events.Dispatch(&EventMuteMe{Duration: 30 * time.Second, User: reward.User.UserName})

	if m.helix != nil {
		_, err := m.helix.UpdateChannelCustomRewardsRedemptionStatus(&helix.UpdateChannelCustomRewardsRedemptionStatusParams{
			ID:            reward.ID,
			BroadcasterID: m.broadcasterID,
			RewardID:      reward.Reward.ID,
			Status:        "FULFILLED",
		})
		if err != nil {
			logger.Errorf("Error updating reward status: %v", err)
		}
	}

	return nil
}

func (m *CalabocaModule) handleMuteMe(e *EventMuteMe) error {
	if m.obs == nil {
		return fmt.Errorf("OBS client not initialized")
	}

	endTime := time.Now().Add(e.Duration)

	inputs, err := m.obs.Inputs.GetInputList(inputRequest.NewGetInputListParams().WithInputKind("pulse_input_capture"))
	if err != nil {
		logger.Errorf("Error getting input list: %s", err)
		return err
	}

	micInput := ""

	for _, input := range inputs.Inputs {
		if strings.Contains(strings.ToLower(input.InputName), "mic") {
			micInput = input.InputName
		}
	}

	_, _ = m.obs.Inputs.SetInputMute(inputRequest.NewSetInputMuteParams().WithInputName(micInput).WithInputMuted(true))

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			tick := <-ticker.C
			diff := endTime.Sub(tick)

			if tick.After(endTime) {
				ticker.Stop()
				_, _ = m.obs.Inputs.SetInputMute(inputRequest.NewSetInputMuteParams().WithInputName(micInput).WithInputMuted(false))

				events.Dispatch(NewWebEvent("force_muted", ""))
				break
			} else {
				events.Dispatch(NewWebEvent("force_muted", views.RenderToString(views.ForceMuted(diff.Round(time.Second).String(), e.User))))
			}
		}
	}()

	return nil
}

var _ events.EventHandler = &CalabocaModule{}
