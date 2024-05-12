package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/zmb3/spotify/v2"
)

type SpotifyModule struct {
	client *spotify.Client
}

func (*SpotifyModule) Register() {
}

func NewSpotifyModule(clientID, clientSecret string) *SpotifyModule {
	return &SpotifyModule{}
}

var _ events.EventHandler = &SpotifyModule{}
