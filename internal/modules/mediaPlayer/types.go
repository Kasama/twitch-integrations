package mediaplayer

import (
	"errors"
	"time"
)

type MediaPlayer interface {
	Play() error
	Pause() error
	Next() error
	Enqueue(query string, requester string, priority Priority) error
	TimeLeft() (time.Duration, error)
	PlayingInfo() (string, error)
}

var ErrNotFound = errors.New("Queried song not found")

type Priority uint8

const (
	PRIORITY_LOW    Priority = 0
	PRIORITY_NORMAL Priority = 1
	PRIORITY_HIGH   Priority = 3
)

type Intent string

const (
	MediaIntentPlay            Intent = "play"
	MediaIntentPause           Intent = "pause"
	MediaIntentNext            Intent = "next"
	MediaIntentEnqueue         Intent = "enqueue"
	MediaIntentPriorityEnqueue Intent = "priority-enqueue"
)

type SongQueueItemType string

const (
	SongQueueItemTypeSpotify = "spotify"
	SongQueueItemTypeYoutube = "youtube"
)

type SongQueueItem struct {
	Type          SongQueueItemType `json:"type"`
	Track         any               `json:"track"`
	User          string            `json:"user"`
	OriginalQuery string            `json:"original_query"`
}

type Event struct {
	Intent       Intent
	EnqueueQuery string
	Requester    string
}

func EnqueueEvent(query string, requester string) *Event {
	return &Event{
		Intent:       MediaIntentEnqueue,
		EnqueueQuery: query,
		Requester:    requester,
	}
}

func PriorityEnqueueEvent(query string, requester string) *Event {
	return &Event{
		Intent:       MediaIntentPriorityEnqueue,
		EnqueueQuery: query,
		Requester:    requester,
	}
}

func PlayEvent() *Event {
	return &Event{
		Intent: MediaIntentPlay,
	}
}

func PauseEvent() *Event {
	return &Event{
		Intent: MediaIntentPause,
	}
}

func NextEvent() *Event {
	return &Event{
		Intent: MediaIntentNext,
	}
}
