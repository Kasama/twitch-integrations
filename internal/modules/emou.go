package modules

import (
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

const rewardIDEmou = "5659b203-3b98-4c9a-a7e8-d945889a8588"
const emouDuration = 1 * time.Minute

type EmouModule struct{}

func NewEmouModule() *EmouModule {
	return &EmouModule{}
}

func (m *EmouModule) Register() {
	events.Register(m.handleReward)
}

func (m *EmouModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDEmou {
		return nil
	}

	endTime := time.Now().Add(emouDuration)

	PlayMp3URL("https://www.myinstants.com/media/sounds/ding_2.mp3")

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			tick := <-ticker.C
			diff := endTime.Sub(tick)

			if tick.After(endTime) {
				ticker.Stop()
				events.Dispatch(NewWebEvent("emou", ""))
				break
			} else {
				events.Dispatch(NewWebEvent("emou", views.RenderToString(views.Emou(diff.Round(time.Second).String(), reward.User.UserName))))
			}
		}
	}()

	return nil
}

var _ events.EventHandler = &EmouModule{}
