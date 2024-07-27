package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	mediaplayer "github.com/Kasama/kasama-twitch-integrations/internal/modules/mediaPlayer"
)

type MediaPlayerModule struct {
	spotifyPlayer *mediaplayer.SpotifyPlayer
	youtubePlayer *mediaplayer.YoutubePlayer
}

func NewMediaPlayerModule() *MediaPlayerModule {
	return &MediaPlayerModule{
		spotifyPlayer: &mediaplayer.SpotifyPlayer{},
		youtubePlayer: &mediaplayer.YoutubePlayer{},
	}
}

// Enqueue implements mediaplayer.MediaPlayer.
func (m *MediaPlayerModule) Enqueue(query string, priority mediaplayer.Priority) error {
	// if strings.Contains(query, "open.spotify.com") {
	// 	m.spotifyPlayer.Enqueue(query, priority)
	// 	// https://open.spotify.com/track/0a7BloCiNzLDD9qSQHh5m7?si=c73c611b98a142ad
	// 	u, err := url.Parse(strings.TrimSpace(query))
	// 	if err != nil {
	// 		return errors.Join(mediaplayer.ErrNotFound, err)
	// 	}
	// 	pathParts := strings.Split(u.Path, "/")
	// 	if len(pathParts) < 2 {
	// 		chatMessage("Link do spotify invalido")
	// 	}
	// 	if pathParts[1] == "track" {
	// 		trackID = pathParts[2]
	// 	}
	// 	if pathParts[2] == "track" {
	// 		trackID = pathParts[3]
	// 	}
	// } else if strings.Contains(query, "youtube.com") || strings.Contains(query, "youtu.be") {
	// 	resp, err := http.Get(strings.TrimSpace(query))
	// 	if err != nil {
	// 		return errors.Join(mediaplayer.ErrNotFound, err)
	// 	}
	// 	document, err := goquery.NewDocumentFromReader(resp.Body)
	// 	if err != nil {
	// 		return errors.Join(mediaplayer.ErrNotFound, err)
	// 	}

	// 	document.Find("meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
	// 		metadata, _ := s.Attr("name")
	// 		if metadata != "title" {
	// 			return true
	// 		}
	// 		title, exists := s.Attr("content")
	// 		if !exists {
	// 			return true
	// 		}

	// 		resp, _ := m.client.Search(m.ctx, title, spotify.SearchTypeTrack)

	// 		if resp.Tracks.Total <= 0 {
	// 			logger.Debugf("SpotifyModule: no tracks found")
	// 			notFoundMessage()
	// 		} else {
	// 			track := resp.Tracks.Tracks[0]
	// 			trackID = track.ID.String()
	// 		}

	// 		return false
	// 	})

	// } else {
	// 	resp, _ := m.client.Search(m.ctx, query, spotify.SearchTypeTrack)

	// 	if resp.Tracks.Total <= 0 {
	// 		logger.Debugf("SpotifyModule: no tracks found")
	// 		notFoundMessage()
	// 	} else {
	// 		track := resp.Tracks.Tracks[0]
	// 		trackID = track.ID.String()
	// 	}
	// }

	// if trackID == "" {
	// 	notFoundMessage()
	// 	return nil
	// }

	// track, _ := m.client.GetTrack(m.ctx, spotify.ID(trackID))
	// if track != nil {
	// 	priority := PRIORITY_NORMAL
	// 	if reward.Reward.ID == songRequestPriorityRewardID {
	// 		priority = PRIORITY_HIGH
	// 	}
	// 	m.queue.Push(priority, SongQueueItem{
	// 		Type:  SongQueueItemTypeSpotify,
	// 		Track: *track,
	// 		User:  reward.UserName,
	// 	})
	// }
	return nil
}

func (m *MediaPlayerModule) Register() {
	events.Register(m.handleEnqueueEvent)
}

var _ events.EventHandler = &MediaPlayerModule{}

func (m *MediaPlayerModule) handleEnqueueEvent(event *mediaplayer.Event) error {
	switch event.Intent {
	case mediaplayer.MediaIntentPlay:
	case mediaplayer.MediaIntentPause:
	case mediaplayer.MediaIntentNext:
	case mediaplayer.MediaIntentEnqueue:
		return m.Enqueue(event.EnqueueQuery, mediaplayer.PRIORITY_NORMAL)
	case mediaplayer.MediaIntentPriorityEnqueue:
		return m.Enqueue(event.EnqueueQuery, mediaplayer.PRIORITY_HIGH)
	}

	return nil
}
