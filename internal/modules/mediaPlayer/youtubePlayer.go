package mediaplayer

import (
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/blang/mpv"
)

const youtubeMediaName = "yt-media-name"

type YoutubePlayer struct {
	player *mpv.Client
}

func NewYoutubePlayer(client *mpv.Client) *YoutubePlayer {
	return &YoutubePlayer{
		player: client,
	}
}

// Enqueue implements MediaPlayer.
func (y *YoutubePlayer) Enqueue(query string, requester string, priority Priority) error {
	_, err := y.player.Exec("loadfile", query, "insert-at-play", -1)
	if err != nil {
		return err
	}

	resp, err := http.Get(strings.TrimSpace(query))
	if err != nil {
		return err
	}
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	err = y.player.SetProperty(youtubeMediaName, "")

	title := ""
	document.Find("meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
		metadata, _ := s.Attr("name")
		if metadata != "title" {
			return true
		}
		var exists bool
		title, exists = s.Attr("content")
		if !exists {
			return true
		}
		err = y.player.SetProperty(youtubeMediaName, title)
		return err != nil
	})

	return err
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
