package modules

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/db"
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/PuerkitoBio/goquery"
	"github.com/gempir/go-twitch-irc/v4"
	twitchEventSub "github.com/joeyak/go-twitch-eventsub/v2"
	"github.com/zmb3/spotify/v2"
)

const (
	PRIORITY_LOW    uint8 = 0
	PRIORITY_NORMAL uint8 = 1
	PRIORITY_HIGH   uint8 = 3
)

// const spotifyQueueName = "spotify_queue"
const spotifyQueueName = "spotify_queue_test"

const songRequestRewardID = "35401b62-32aa-4009-ac1e-a5f3015670e8"
const songRequestPriorityRewardID = "ee3b013d-619b-45d8-8745-71c33bf71e6b"
const skipSongRewardID = "b062e370-6a64-4537-b513-f83bd1588496"

const (
	SongQueueItemTypeSpotify = "spotify"
	SongQueueItemTypeYoutube = "youtube"
)

type SongQueueItem struct {
	Type  string            `json:"type"`
	Track spotify.FullTrack `json:"track"`
	User  string            `json:"user"`
}

type playingState struct {
	nowPlaying   *SongQueueItem
	lastDequeued *SongQueueItem
	locked       atomic.Bool
}

type SpotifyModule struct {
	ctx              context.Context
	client           *spotify.Client
	twitchChatClient *twitch.Client
	twitchChannel    string
	queue            *db.Queue[SongQueueItem]
	playingState     playingState
}

func (m *SpotifyModule) Register() {
	events.Register(m.handleSpotifyClient)
	events.Register(m.handleTwtichChatClient)
	events.Register(m.handlePlayMessage)
	events.Register(m.handlePauseMessage)
	events.Register(m.handleSongRequestReward)
	events.Register(m.handleSkipTrackReward)
	events.Register(m.handleMusicInfo)
	events.Register(m.handleTimer)
}

func NewSpotifyModule(ctx context.Context, twitchUsername string) *SpotifyModule {
	q, err := db.NewQueue[SongQueueItem](spotifyQueueName)
	if err != nil {
		logger.Errorf("SpotifyModule: error creating queue: %v", err)
		return nil
	}

	logger.Debugf("SpotifyModule: created queue")
	events.Dispatch(q)
	logger.Debugf("SpotifyModule: dispatched queue")

	return &SpotifyModule{
		ctx:           ctx,
		client:        nil,
		twitchChannel: twitchUsername,
		queue:         q,
		playingState: playingState{
			nowPlaying:   &SongQueueItem{},
			lastDequeued: nil,
			locked:       atomic.Bool{}, // default is false
		},
	}
}

func (m *SpotifyModule) Queue() *db.Queue[SongQueueItem] {
	return m.queue
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
	if reward.Reward.ID != songRequestRewardID && reward.Reward.ID != songRequestPriorityRewardID {
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
		priority := PRIORITY_NORMAL
		if reward.Reward.ID == songRequestPriorityRewardID {
			priority = PRIORITY_HIGH
		}
		m.queue.Push(priority, SongQueueItem{
			Type:  SongQueueItemTypeSpotify,
			Track: *track,
			User:  reward.UserName,
		})

		// _ = m.client.QueueSong(m.ctx, track.ID)

		if m.twitchChatClient != nil {
			m.twitchChatClient.Say(m.twitchChannel, fmt.Sprintf("Adicionado \"%s - %s\" na fila. %d músicas na fila.", track.Artists[0].Name, track.Name, m.queue.Len()))
		}
	}

	return nil
}

func (m *SpotifyModule) handleMusicInfo(msg *twitch.PrivateMessage) error {
	if m.client == nil {
		return nil
	}
	if msg.Message == "!musica" {
		currentlyPlaying, err := m.client.PlayerCurrentlyPlaying(m.ctx)
		if err != nil {
			return err
		}
		if !currentlyPlaying.Playing {
			m.twitchChatClient.Say(m.twitchChannel, "Nenhuma musica tocando no momento")
			return nil
		}
		if m.twitchChatClient != nil {
			next := m.queue.Peek()
			nextText := ""
			if next != nil {
				nextText = fmt.Sprintf(" Proxima: '%s - %s'", next.Track.Artists[0].Name, next.Track.Name)
			}
			queuedTracks := m.queue.Len()
			logger.Debugf("Currently playing: '%s'. %d tracks in queue", currentlyPlaying.Item, queuedTracks)
			m.twitchChatClient.Say(m.twitchChannel, fmt.Sprintf("Tocando: '%s - %s'. Temos mais %d na fila.%s", currentlyPlaying.Item.Artists[0].Name, currentlyPlaying.Item.Name, queuedTracks, nextText))
		}
	}
	return nil
}

func (m *SpotifyModule) handleSkipTrackReward(reward *twitchEventSub.EventChannelChannelPointsCustomRewardRedemptionAdd) error {
	if reward.Reward.ID != skipSongRewardID {
		return nil
	}
	if m.client == nil {
		return nil
	}

	err := m.enqueueNext()
	if err != nil {
		return err
	}

	return m.client.Next(m.ctx)
}

func (m *SpotifyModule) enqueueNext() error {
	if m.client == nil {
		return nil
	}

	if m.queue.Len() == 0 {
		return nil
	}

	spotifyQueue, err := m.client.GetQueue(m.ctx)
	if err != nil || len(spotifyQueue.Items) < 1 {
		return err
	}

	if m.playingState.nowPlaying.Track.ID != spotifyQueue.CurrentlyPlaying.SimpleTrack.ID {
		logger.Debugf("SpotifyModule: Current song is different from spotify queue")
		return nil
	}

	nextSongInSpotifyQueue := spotifyQueue.Items[0]

	if m.playingState.lastDequeued != nil {
		if nextSongInSpotifyQueue.ID == m.playingState.lastDequeued.Track.ID {
			logger.Debugf("SpotifyModule: ignoring request to enqueue %s. Already enqueued", nextSongInSpotifyQueue.Name)
			return nil
		}
	}

	nextSong := m.queue.Pop()
	m.playingState.lastDequeued = nextSong
	if nextSong != nil {
		return m.client.QueueSong(m.ctx, nextSong.Track.ID)
	} else {
		return nil
	}
}

func (m *SpotifyModule) handleTimer(t *time.Time) error {
	if m.client == nil || m.ctx == nil {
		return nil
	}

	currentlyPlaying, err := m.client.PlayerCurrentlyPlaying(m.ctx)
	if err != nil {
		return err
	}

	if currentlyPlaying.Playing {
		if m.playingState.nowPlaying.Track.ID != currentlyPlaying.Item.ID {
			events.Dispatch(NewWebEvent("music_now_playing", views.RenderToString(views.MusicNowPlaying(currentlyPlaying.Item.Artists[0].Name, currentlyPlaying.Item.Name, m.playingState.nowPlaying.User))))
		}
		m.playingState.nowPlaying.Track = *currentlyPlaying.Item
		if m.playingState.lastDequeued != nil {
			if currentlyPlaying.Item.ID == m.playingState.lastDequeued.Track.ID {
				m.playingState.nowPlaying.User = m.playingState.lastDequeued.User
			} else {
				m.playingState.nowPlaying.User = ""
			}
		}
	} else {
		m.playingState.nowPlaying = &SongQueueItem{}
	}

	triggerTime := 5 * time.Second
	if !currentlyPlaying.Playing || int64(currentlyPlaying.Item.Duration-currentlyPlaying.Progress) > triggerTime.Milliseconds() {
		return nil
	}

	return m.enqueueNext()
}

func (m *SpotifyModule) handlePlayMessage(msg *twitch.PrivateMessage) error {
	if m.client == nil {
		return nil
	}
	if _, ok := msg.User.Badges["broadcaster"]; !ok {
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
