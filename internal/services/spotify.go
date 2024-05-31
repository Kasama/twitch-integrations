package services

import (
	"context"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	spotifyInternal "github.com/Kasama/kasama-twitch-integrations/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type EventSpotifyPlay struct {
	onlyIfPausedByEvent bool
}

func NewEventSpotifyPlay(onlyIfPausedByEvent bool) *EventSpotifyPlay {
	return &EventSpotifyPlay{
		onlyIfPausedByEvent: onlyIfPausedByEvent,
	}
}

type EventSpotifyPause struct{}

func NewEventSpotifyPause() *EventSpotifyPause {
	return &EventSpotifyPause{}
}

type SpotifyService struct {
	client        *spotify.Client
	config        *spotifyInternal.SpotifyConfig
	ctx           context.Context
	pausedByEvent bool
}

func NewSpotifyService(ctx context.Context, config *spotifyInternal.SpotifyConfig) *SpotifyService {
	return &SpotifyService{
		client:        nil,
		config:        config,
		ctx:           ctx,
		pausedByEvent: false,
	}
}

// Register implements events.EventHandler.
func (s *SpotifyService) Register() {
	events.Register(s.handleToken)
	events.Register(s.handlePlayEvent)
	events.Register(s.handlePauseEvent)
}

func (s *SpotifyService) handleToken(token *spotifyInternal.SpotifyAuth) error {
	httpClient := s.config.Oauth2config.Client(s.ctx, token.Token)
	client := spotify.New(httpClient)

	s.client = client

	user, err := client.CurrentUser(s.ctx)
	if err != nil {
		return err
	}

	logger.Debugf("spotify connected with account name '%s'", user.DisplayName)

	events.Dispatch(client)

	return nil
}

func (s *SpotifyService) handlePlayEvent(event *EventSpotifyPlay) error {
	if s.client == nil {
		return nil
	}
	if event.onlyIfPausedByEvent && !s.pausedByEvent {
		return nil
	}
	s.pausedByEvent = false

	err := s.client.Play(s.ctx)

	return err
}

func (s *SpotifyService) handlePauseEvent(event *EventSpotifyPause) error {
	if s.client == nil {
		return nil
	}

	logger.Debugf("SpotifyService: received pause event")
	state, err := s.client.PlayerState(s.ctx)
	if err != nil {
		return err
	}
	logger.Debugf("SpotifyService: state is %v", state)

	if state.CurrentlyPlaying.Playing || true {
		logger.Debugf("Trying to pause")
		err = s.client.Pause(s.ctx)
		if err != nil {
			return err
		}
		s.pausedByEvent = true
	}

	return nil
}

var _ events.EventHandler = &SpotifyService{}
