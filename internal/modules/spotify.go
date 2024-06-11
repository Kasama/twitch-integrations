package modules

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/PuerkitoBio/goquery"
	"github.com/gempir/go-twitch-irc/v4"
	twitchEventSub "github.com/joeyak/go-twitch-eventsub/v2"
	"github.com/zmb3/spotify/v2"
)

const songRequestRewardID = "35401b62-32aa-4009-ac1e-a5f3015670e8"

type SpotifyModule struct {
	ctx              context.Context
	client           *spotify.Client
	twitchChatClient *twitch.Client
	twitchChannel    string
}

func (m *SpotifyModule) Register() {
	events.Register(m.handleSpotifyClient)
	events.Register(m.handleTwtichChatClient)
	events.Register(m.handlePlayMessage)
	events.Register(m.handlePauseMessage)
	events.Register(m.handleSongRequestReward)
}

func NewSpotifyModule(ctx context.Context, twitchUsername string) *SpotifyModule {
	return &SpotifyModule{
		ctx:           ctx,
		client:        nil,
		twitchChannel: twitchUsername,
	}
}

func (m *SpotifyModule) handleTwtichChatClient(client *twitch.Client) error {
	m.twitchChatClient = client
	return nil
}

func (m *SpotifyModule) handleSpotifyClient(client *spotify.Client) error {
	m.client = client
	return nil
}

func (m *SpotifyModule) handleSongRequestReward(reward *twitchEventSub.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != songRequestRewardID {
		return nil
	}
	if m.client == nil {
		return nil
	}

	chatMessage := func(msg string) {
		if m.twitchChatClient != nil {
			m.twitchChatClient.Say(m.twitchChannel, msg)
		}
	}
	notFoundMessage := func() {
		chatMessage("Musica não encontrada")
	}

	query := reward.UserInput
	var trackID string

	if strings.Contains(query, "open.spotify.com") {
		// https: //open.spotify.com/track/0a7BloCiNzLDD9qSQHh5m7?si=c73c611b98a142ad
		u, err := url.Parse(strings.TrimSpace(query))
		if err != nil {
			notFoundMessage()
			return err
		}
		pathParts := strings.Split(u.Path, "/")
		if len(pathParts) < 2 {
			chatMessage("Link do spotify invalido")
		}
		if pathParts[1] == "track" {
			trackID = pathParts[2]
		}
		if pathParts[2] == "track" {
			trackID = pathParts[3]
		}
	} else if strings.Contains(query, "youtube.com") || strings.Contains(query, "youtu.be") {
		resp, err := http.Get(strings.TrimSpace(query))
		if err != nil {
			notFoundMessage()
			return err
		}
		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			notFoundMessage()
			return err
		}

		document.Find("meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
			metadata, _ := s.Attr("name")
			if metadata != "title" {
				return true
			}
			title, exists := s.Attr("content")
			if !exists {
				return true
			}

			resp, _ := m.client.Search(m.ctx, title, spotify.SearchTypeTrack)

			if resp.Tracks.Total <= 0 {
				logger.Debugf("SpotifyModule: no tracks found")
				notFoundMessage()
			} else {
				track := resp.Tracks.Tracks[0]
				trackID = track.ID.String()
			}

			return false
		})

	} else {
		resp, _ := m.client.Search(m.ctx, query, spotify.SearchTypeTrack)

		if resp.Tracks.Total <= 0 {
			logger.Debugf("SpotifyModule: no tracks found")
			notFoundMessage()
		} else {
			track := resp.Tracks.Tracks[0]
			trackID = track.ID.String()
		}
	}

	if trackID == "" {
		notFoundMessage()
		return nil
	}

	track, _ := m.client.GetTrack(m.ctx, spotify.ID(trackID))
	if track != nil {
		if m.twitchChatClient != nil {
			m.twitchChatClient.Say(m.twitchChannel, fmt.Sprintf("Adicionado \"%s - %s\" na fila", track.Artists[0].Name, track.Name))
		}
	}
	_ = m.client.QueueSong(m.ctx, spotify.ID(trackID))

	return nil
}

func (m *SpotifyModule) handlePlayMessage(msg *twitch.PrivateMessage) error {
	if m.client == nil {
		return nil
	}

	logger.Debugf("SpotifyModule: message '%s'", msg.Message)
	if !strings.HasPrefix(msg.Message, "!play") {
		return nil
	}
	query := strings.TrimPrefix(msg.Message, "!play ")

	resp, _ := m.client.Search(m.ctx, query, spotify.SearchTypeTrack)

	if resp.Tracks.Total <= 0 {
		logger.Debugf("SpotifyModule: no tracks found")
		if m.twitchChatClient != nil {
			m.twitchChatClient.Say(m.twitchChannel, "Musica não encontrada")
		}
	} else {
		track := resp.Tracks.Tracks[0]
		_ = m.client.QueueSong(m.ctx, track.ID)
		if m.twitchChatClient != nil {
			m.twitchChatClient.Say(m.twitchChannel, fmt.Sprintf("Adicionado %s - %s na fila", track.Artists[0].Name, track.Name))
		}
	}

	logger.Debugf("SpotifyModule: received play command")

	return nil
}

func (m *SpotifyModule) handlePauseMessage(msg *twitch.PrivateMessage) error {
	if msg.Message != "!pause" {
		return nil
	}

	logger.Debugf("SpotifyModule: received pause command")

	err := m.client.Pause(context.Background())

	logger.Debugf("Pausing err: %v", err)

	return nil
}

var _ events.EventHandler = &SpotifyModule{}
