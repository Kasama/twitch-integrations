package modules

import (
	"math/rand"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

const floresbertoSoundRewardID = "c44fd675-4d27-46c5-a8a2-3fe05130ca07"

type FloresbertoModule struct {
}

func NewFloresbertoModule() *FloresbertoModule {
	return &FloresbertoModule{}
}

// Register implements events.EventHandler.
func (m *FloresbertoModule) Register() {
	events.Register(m.handleReward)
}

func (m *FloresbertoModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != floresbertoSoundRewardID {
		return nil
	}

	url := ""
	if rand.Intn(2) == 0 {
		url = "https://www.myinstants.com/media/sounds/floresberto-perdendo-tempo.mp3"
	} else {
		url = "https://www.myinstants.com/media/sounds/floresberto-ganhando-kill.mp3"
	}

	resp, err := GetMp3Reader(url)
	if err != nil {
		return err
	}

	events.Dispatch(NewPlayAudioEvent(resp, true))

	return nil
}

var _ events.EventHandler = &FloresbertoModule{}
