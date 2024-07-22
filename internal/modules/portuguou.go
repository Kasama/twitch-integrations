package modules

import (
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

const rewardIDPortugou = "7af7f6e9-47b9-4800-a6b8-1998f1a1fbdf"
const portugouDuration = 1 * time.Minute

type PortugouModule struct{}

func NewPortugouModule() *PortugouModule {
	return &PortugouModule{}
}

func (m *PortugouModule) Register() {
	events.Register(m.handleReward)
}

func (m *PortugouModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDPortugou {
		return nil
	}

	endTime := time.Now().Add(portugouDuration)

	PlayMp3URL("https://www.myinstants.com/media/sounds/ding_2.mp3")

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			tick := <-ticker.C
			diff := endTime.Sub(tick)

			if tick.After(endTime) {
				ticker.Stop()
				events.Dispatch(NewWebEvent("portugou", ""))
				break
			} else {
				events.Dispatch(NewWebEvent("portugou", views.RenderToString(views.Portugou(diff.Round(time.Second).String(), reward.User.UserName))))
			}
		}
	}()

	return nil
}

var _ events.EventHandler = &PortugouModule{}
