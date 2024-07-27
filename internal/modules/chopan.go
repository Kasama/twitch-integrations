package modules

import (
	"encoding/base64"
	"io/fs"
	"math/rand"
	"os"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	obsEvents "github.com/andreykaipov/goobs/api/events"
	"github.com/gempir/go-twitch-irc/v4"
	twitchEvents "github.com/joeyak/go-twitch-eventsub/v2"
)

const pharseDir = "/home/roberto/documents/programming/Twitch/kasama-twitch-integrations/assets/chopan-phrases/"
const currentPhrase = "/tmp/currentChopanPhrase"
const chopanRewardID = "b76fe0e6-48c4-40cf-91e6-990cad1f7217"

type ChopanModule struct {
	imageBlob string
}

func NewChopanModule() *ChopanModule {
	imageBlob := ""
	_, err := os.Stat(currentPhrase)
	if err == nil {
		phraseBlob, err := os.ReadFile(currentPhrase)
		if err == nil {
			base64blob := base64.StdEncoding.EncodeToString(phraseBlob)
			imageBlob = "data:image/webp;base64," + base64blob
		}
	}

	return &ChopanModule{
		imageBlob: imageBlob,
	}
}

// Register implements events.EventHandler.
func (m *ChopanModule) Register() {
	events.Register(m.handleReward)
	events.Register(m.handleResetPhrase)
	events.Register(m.handleStartStream)
}

func (m *ChopanModule) resetPhrase() (string, error) {
	dir := os.DirFS(pharseDir)

	phrases, err := fs.Glob(dir, "*.webp")
	if err != nil {
		return "", err
	}

	phrase := rand.Intn(len(phrases))

	phraseBlob, err := fs.ReadFile(dir, phrases[phrase])
	if err != nil {
		return "", err
	}

	base64blob := base64.StdEncoding.EncodeToString(phraseBlob)
	_ = os.WriteFile(currentPhrase, phraseBlob, 0644)

	return "data:image/webp;base64," + base64blob, nil
}

func (m *ChopanModule) handleResetPhrase(msg *twitch.PrivateMessage) error {
	if msg.User.Badges["broadcaster"] != 1 || (msg.Message != "!reset-phrase" && msg.Message != "!chopan") {
		return nil
	}

	if msg.Message != "!chopan" {
		imageBlob, err := m.resetPhrase()
		if err != nil {
			return err
		}
		m.imageBlob = imageBlob
	}

	events.Dispatch(NewWebEvent("chopan_phrase", views.RenderToString(views.ChopanPhrase(m.imageBlob))))

	return nil
}

func (m *ChopanModule) handleStartStream(e *obsEvents.StreamStateChanged) error {

	imageBlob, err := m.resetPhrase()
	m.imageBlob = imageBlob

	return err
}

func (m *ChopanModule) handleReward(reward *twitchEvents.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != chopanRewardID {
		return nil
	}

	events.Dispatch(NewWebEvent("chopan_phrase", views.RenderToString(views.ChopanPhrase(m.imageBlob))))

	return nil
}

var _ events.EventHandler = &ChopanModule{}
