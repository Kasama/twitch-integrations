package mediaplayer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/zmb3/spotify/v2"
)

var ErrInvalidSpotifyLink = errors.New("invalid spotify link")

type playerState struct {
	playing bool
}

type SpotifyPlayer struct {
	client         *spotify.Client
	lastKnownState playerState
	ctx            context.Context
}

func NewSpotifyPlayer(ctx context.Context, client *spotify.Client) *SpotifyPlayer {
	return &SpotifyPlayer{
		ctx:            ctx,
		client:         client,
		lastKnownState: playerState{playing: false},
	}
}

// Enqueue implements MediaPlayer.
func (s *SpotifyPlayer) Enqueue(query string, requester string, priority Priority) error {
	trackID := ""
	if strings.Contains(query, "open.spotify.com") {
		// https://open.spotify.com/track/0a7BloCiNzLDD9qSQHh5m7?si=c73c611b98a142ad
		u, err := url.Parse(strings.TrimSpace(query))
		if err != nil {
			return err
		}
		pathParts := strings.Split(u.Path, "/")
		if len(pathParts) < 2 {
			return ErrInvalidSpotifyLink
		}
		if pathParts[1] == "track" {
			trackID = pathParts[2]
		}
		if pathParts[2] == "track" {
			trackID = pathParts[3]
		}
	}

	track, err := s.client.GetTrack(s.ctx, spotify.ID(trackID))
	if err != nil {
		return err
	}
	return s.client.QueueSong(s.ctx, track.ID)
}

// Next implements MediaPlayer.
func (s *SpotifyPlayer) Next() error {
	return s.client.Next(s.ctx)
}

// Pause implements MediaPlayer.
func (s *SpotifyPlayer) Pause() error {
	return s.client.Pause(s.ctx)
}

// Play implements MediaPlayer.
func (s *SpotifyPlayer) Play() error {
	return s.client.Play(s.ctx)
}

// PlayingInfo implements MediaPlayer.
func (s *SpotifyPlayer) PlayingInfo() (string, error) {
	currentlyPlaying, err := s.client.PlayerCurrentlyPlaying(s.ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s - %s", currentlyPlaying.Item.Artists[0].Name, currentlyPlaying.Item.Name), nil
}

// TimeLeft implements MediaPlayer.
func (s *SpotifyPlayer) TimeLeft() (time.Duration, error) {
	currentlyPlaying, err := s.client.PlayerCurrentlyPlaying(s.ctx)
	if err != nil {
		return 0, err
	}
	milisRemaining := (currentlyPlaying.Item.Duration-currentlyPlaying.Progress)
	return time.Duration(milisRemaining) * time.Millisecond, nil
}

var _ MediaPlayer = &SpotifyPlayer{}
