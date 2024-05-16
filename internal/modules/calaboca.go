package modules

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/andreykaipov/goobs"
	inputRequest "github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

const rewardIDMuteMe = "1bbe8c0f-d7f5-48f2-a658-7b6ad2a2546d"
const micMutedFile = "/tmp/mic-muted-forced.txt"
const micMutedOwnerFile = "/tmp/mic-muted-forced-cause.txt"

type EventMuteMe struct {
	Duration time.Duration
	User     string
}

type CalabocaModule struct {
	obs *goobs.Client
}

func NewCalabocaModule() *CalabocaModule {
	return &CalabocaModule{}
}

func (m *CalabocaModule) Register() {
	_ = os.WriteFile(micMutedFile, []byte(""), 0644)
	_ = os.WriteFile(micMutedOwnerFile, []byte(""), 0644)

	events.Register(m.handleOBSClient)
	events.Register(m.handleReward)
	events.Register(m.handleMuteMe)
}

func (m *CalabocaModule) handleOBSClient(client *goobs.Client) error {
	m.obs = client

	return nil
}

func (m *CalabocaModule) handleReward(reward *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != rewardIDMuteMe {
		return nil
	}
	events.Dispatch(&EventMuteMe{Duration: 30 * time.Second, User: reward.User.UserName})
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
			_ = os.WriteFile(micMutedFile, []byte(fmt.Sprintf("  por %s", diff.Round(time.Second).String())), 0644)
			_ = os.WriteFile(micMutedOwnerFile, []byte(e.User), 0644)

			if tick.After(endTime) {
				ticker.Stop()
				_, _ = m.obs.Inputs.SetInputMute(inputRequest.NewSetInputMuteParams().WithInputName(micInput).WithInputMuted(false))
				_ = os.WriteFile(micMutedFile, []byte(""), 0644)
				_ = os.WriteFile(micMutedOwnerFile, []byte(""), 0644)
				break
			}
		}
	}()

	return nil
}

var _ events.EventHandler = &CalabocaModule{}
