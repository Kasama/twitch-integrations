package modules

import (
	"io"
	"net/http"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	services "github.com/Kasama/kasama-twitch-integrations/internal/services"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gempir/go-twitch-irc/v4"
)

const usualMP3SampleRate beep.SampleRate = beep.SampleRate(44100) // 44.1KHz
const resamplingQuality = 4

type PlayAudioEvent struct {
	Reader      io.ReadCloser
	pausesMusic bool
}

func NewPlayAudioEvent(reader io.ReadCloser, pausesMusic bool) *PlayAudioEvent {
	return &PlayAudioEvent{
		Reader:      reader,
		pausesMusic: pausesMusic,
	}
}

func GetMp3Reader(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func PlayMp3URL(url string) {
	resp, err := GetMp3Reader(url)
	if err != nil {
		return
	}

	events.Dispatch(NewPlayAudioEvent(resp, false))
}

type AudioModule struct {
	done chan bool
	stop chan struct{}
}

func NewAudioModule() *AudioModule {
	return &AudioModule{
		done: nil,
		stop: nil,
	}
}

func (m *AudioModule) Register() {
	events.Register(m.handleCommand)
	events.Register(m.handleStop)
	events.Register(m.handlePlayAudio)

	err := speaker.Init(usualMP3SampleRate, usualMP3SampleRate.N(time.Second/10))
	if err != nil {
		logger.Errorf("Failed to start speaker with sample rate %fKHz: %s", float32(usualMP3SampleRate)/1000, err.Error())
	}
}

func (m *AudioModule) handlePlayAudio(event *PlayAudioEvent) error {
	if m.stop != nil {
		m.stop <- struct{}{}
		m.stop = nil
	}

	m.done = make(chan bool)
	m.stop = make(chan struct{})

	go func() {
		streamer, format, err := mp3.Decode(event.Reader)
		if err != nil {
			logger.Errorf("Failed to decode music: %s", err.Error())
		}
		defer streamer.Close()

		resampled := beep.Resample(resamplingQuality, format.SampleRate, usualMP3SampleRate, streamer)

		if event.pausesMusic {
			events.Dispatch(services.NewEventSpotifyPause())
		}
		speaker.Play(beep.Seq(resampled, beep.Callback(func() {
			if m.done != nil {
				m.done <- true
				m.done = nil
			}
			if event.pausesMusic {
				events.Dispatch(services.NewEventSpotifyPlay(true))
			}
		})))

		go func() {
			timer := time.NewTimer(10 * time.Second)
			select {
			case <-timer.C:
				if m.stop != nil {
					m.stop <- struct{}{}
					if event.pausesMusic {
						events.Dispatch(services.NewEventSpotifyPlay(true))
					}
				}
				break
			case <-m.done:
				break
			case <-m.stop:
				break
			}
		}()

		select {
		case <-m.done:
			m.done = nil
			m.stop = nil
			break
		case <-m.stop:
			speaker.Clear()
			m.done = nil
			m.stop = nil
			break
		}
	}()

	return nil
}

func (m *AudioModule) handleStop(message *twitch.PrivateMessage) error {
	if message.Message != "!stop" {
		return nil
	}

	if m.stop != nil {
		m.stop <- struct{}{}
		m.stop = nil
	}

	return nil
}

func (m *AudioModule) handleCommand(message *twitch.PrivateMessage) error {
	if message.Message != "!play-nossa-test" {
		return nil
	}

	if m.stop != nil {
		m.stop <- struct{}{}
		m.stop = nil
	}

	m.done = make(chan bool)
	m.stop = make(chan struct{})

	go func() {
		logger.Debugf("open music")
		// music, err := os.Open("/home/roberto/Music/A Cruel Angel's Thesis.mp3")
		// if err != nil {
		// 	// return err
		// }
		// defer music.Close()

		resp, _ := http.Get("https://www.myinstants.com/media/sounds/seu-madruga-nossa.mp3")

		logger.Debugf("decode music")
		streamer, format, err := mp3.Decode(resp.Body)
		if err != nil {
			// return err
		}
		defer streamer.Close()

		resampled := beep.Resample(resamplingQuality, format.SampleRate, usualMP3SampleRate, streamer)

		speaker.Play(beep.Seq(resampled, beep.Callback(func() {
			logger.Debugf("music done")
			if m.done != nil {
				m.done <- true
				m.done = nil
			}
		})))

		logger.Debugf("waiting for something")
		select {
		case <-m.done:
			logger.Debugf("music got finished")
			break
		case <-m.stop:
			logger.Debugf("music was stopped")
			speaker.Clear()
			break
		}
	}()

	return nil
}

var _ events.EventHandler = &AudioModule{}
