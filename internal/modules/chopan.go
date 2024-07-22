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
const chopanRewardID = "b76fe0e6-48c4-40cf-91e6-990cad1f7217"

type ChopanModule struct {
	imageBlob string
}

func NewChopanModule() *ChopanModule {
	return &ChopanModule{
		imageBlob: "",
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

	return "data:image/webp;base64," + base64blob, nil
}

func (m *ChopanModule) handleResetPhrase(msg *twitch.PrivateMessage) error {
	if msg.User.Badges["broadcaster"] != 1 || msg.Message != "!reset-phrase" {
		return nil
	}

	imageBlob, err := m.resetPhrase()
	m.imageBlob = imageBlob

	events.Dispatch(NewWebEvent("chopan_phrase", views.RenderToString(views.ChopanPhrase(m.imageBlob))))

	return err
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
