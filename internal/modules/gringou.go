package modules

import (
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

const rewardIDGringou = "3ae88181-a953-4935-b34c-ee5332b88d5d"
const gringouDuration = 1 * time.Minute

type GringouModule struct{}

func NewGringouModule() *GringouModule {
	return &GringouModule{}
}

func (m *GringouModule) Register() {
	events.Register(m.handleReward)
}

func (m *GringouModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDGringou {
		return nil
	}

	events.Dispatch(NewPunishableRedeemInfo(reward.User.UserID, reward.User.UserName))
	endTime := time.Now().Add(gringouDuration)

	PlayMp3URL("https://www.myinstants.com/media/sounds/ding_2.mp3")

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			tick := <-ticker.C
			diff := endTime.Sub(tick)

			if tick.After(endTime) {
				ticker.Stop()
				events.Dispatch(NewWebEvent("gringou", ""))
				break
			} else {
				events.Dispatch(NewWebEvent("gringou", views.RenderToString(views.Gringou(diff.Round(time.Second).String(), reward.User.UserName))))
			}
		}
	}()

	return nil
}

var _ events.EventHandler = &GringouModule{}
