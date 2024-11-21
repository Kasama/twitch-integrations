package modules

import (
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/general"
	"github.com/gempir/go-twitch-irc/v4"
)

type ReplayModule struct {
	obsClient *goobs.Client
	lastUsed  time.Time
}

func NewReplayModule() *ReplayModule {
	zero := time.Unix(0, 0)
	return &ReplayModule{
		lastUsed: zero,
	}
}

func (m *ReplayModule) Register() {
	events.Register(m.handleMessage)
	events.Register(m.handleOBSClient)
}

func (m *ReplayModule) handleOBSClient(client *goobs.Client) error {
	m.obsClient = client
	return nil
}

func (m *ReplayModule) handleMessage(msg *twitch.PrivateMessage) error {
	if msg.Message != "!replay" {
		return nil
	}

	if m.lastUsed.Add(time.Minute * 2).After(time.Now()) {
		events.Dispatch(NewChatMessage("Replay usado a menos de 2 minutos"))
		return nil
	}
	m.lastUsed = time.Now()

	hotkeyName := "instant_replay.trigger"
	resp, err := m.obsClient.General.TriggerHotkeyByName(&general.TriggerHotkeyByNameParams{
		HotkeyName: &hotkeyName,
	})
	if err != nil {
		logger.Debugf("Error triggering hotkey: %s", err)
	} else {
		logger.Debugf("Triggered hotkey: %+v", resp)
	}

	return nil
}
