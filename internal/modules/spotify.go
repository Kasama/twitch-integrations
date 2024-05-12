package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/zmb3/spotify/v2"
)

type SpotifyModule struct {
	client *spotify.Client
}

func NewSpotifyModule(clientID, clientSecret string) *SpotifyModule {
	return &SpotifyModule{ }
}

func (*SpotifyModule) HandleEvent(ctx *events.EventContext, event *events.Event) error {
	panic("unimplemented")
}

func (*SpotifyModule) ShouldHandle(ctx *events.EventContext, event *events.Event) bool {
	return event.Kind == events.EventKindChannelPointsRewardRedemption && event.ChannelPointsRewardRedemption.Reward.Title == "Pede uma musica"
}

var _ events.EventHandler = &SpotifyModule{}
