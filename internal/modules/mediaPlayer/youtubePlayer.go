package mediaplayer

import (
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/db"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/blang/mpv"
)

const youtubeMediaName = "yt-media-name"

type YoutubePlayer struct {
	player          *mpv.Client
	defaultPlaylist *db.Queue[string]
	requestQueue    *db.Queue[string]
	stopChan        chan struct{}
}

func NewYoutubePlayer(client *mpv.Client, defaultPlaylistUrls []string) (*YoutubePlayer, error) {
	defaultPlaylist, err := db.NewQueue[string]("default_playlist")
	if err != nil {
		return nil, err
	}

	requestQueue, err := db.NewQueue[string]("request_queue")
	if err != nil {
		return nil, err
	}

	player := &YoutubePlayer{
		player:          client,
		defaultPlaylist: defaultPlaylist,
		requestQueue:    requestQueue,
		stopChan:        make(chan struct{}),
	}

	// Initialize default playlist if it's empty
	if player.defaultPlaylist.Len() == 0 {
		for _, url := range defaultPlaylistUrls {
			player.defaultPlaylist.Push(0, url)
		}
	}

	// Start the playback monitor
	go player.monitorPlayback()

	return player, nil
}

func (y *YoutubePlayer) loadNextTrack() error {
	// First try to play from request queue
	if y.requestQueue.Len() > 0 {
		if nextTrack := y.requestQueue.Pop(); nextTrack != nil {
			_, err := y.player.Exec("loadfile", *nextTrack, "replace")
			return err
		}
	}

	// If no requests, play from default playlist
	if y.defaultPlaylist.Len() > 0 {
		nextTrack := y.defaultPlaylist.Pop()
		if nextTrack != nil {
			// Push it back to maintain rotation
			y.defaultPlaylist.Push(0, *nextTrack)
			_, err := y.player.Exec("loadfile", *nextTrack, "replace")
			return err
		}
	}

	return nil
}

func (y *YoutubePlayer) monitorPlayback() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-y.stopChan:
			return
		case <-ticker.C:
			// Check if we reached the end of the current file
			eofReached, err := y.player.GetBoolProperty("eof-reached")
			if err != nil {
				continue
			}

			// Check if player is idle
			idle, err := y.isIdle()
			logger.Debugf("play idle %v, err: %v", idle, err)
			if err != nil {
				continue
			}

			if eofReached {
				logger.Debugf("Loading next track. EOF: %v, idle %v", eofReached, idle)
				y.loadNextTrack()
			}
		}
	}
}

func (y *YoutubePlayer) isIdle() (bool, error) {
	return y.player.GetBoolProperty("core-idle")
}

// Enqueue implements MediaPlayer.
func (y *YoutubePlayer) Enqueue(query string, requester string, priority Priority) error {
	var queuePriority uint8
	if priority == PRIORITY_HIGH {
		queuePriority = 2
	} else {
		queuePriority = 1
	}

	// Add to request queue
	y.requestQueue.Push(queuePriority, query)

	// If player is idle, start playing immediately
	isIdle, err := y.isIdle()
	if err == nil && isIdle {
		return y.loadNextTrack()
	}

	return nil
}

func (y *YoutubePlayer) AddToDefaultPlaylist(url string) {
	y.defaultPlaylist.Push(0, url)
}

func (y *YoutubePlayer) ClearDefaultPlaylist() error {
	return y.defaultPlaylist.Clear()
}

// Next implements MediaPlayer.
func (y *YoutubePlayer) Next() error {
	_, err := y.player.Exec("playlist-next", "force")
	return err
}

// Pause implements MediaPlayer.
func (y *YoutubePlayer) Pause() error {
	return y.player.SetPause(true)
}

// Play implements MediaPlayer.
func (y *YoutubePlayer) Play() error {
	return y.player.SetPause(false)
}

func (y *YoutubePlayer) PlayPause() error {
	paused, err := y.player.Pause()
	if err != nil {
		return err
	}
	err = y.player.SetPause(!paused)
	if err != nil {
		return err
	}
	return nil
}

func (y *YoutubePlayer) Stop() {
	close(y.stopChan)
}

// PlayingInfo implements MediaPlayer.
func (y *YoutubePlayer) PlayingInfo() (string, error) {
	title, err := y.player.GetProperty(youtubeMediaName)
	if err != nil {
		return "", err
	}
	return title, nil
}

// TimeLeft implements MediaPlayer.
func (y *YoutubePlayer) TimeLeft() (time.Duration, error) {
	panic("unimplemented")
}

var _ MediaPlayer = &YoutubePlayer{}
