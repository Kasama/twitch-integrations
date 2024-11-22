package modules

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	mediaplayer "github.com/Kasama/kasama-twitch-integrations/internal/modules/mediaPlayer"
	services "github.com/Kasama/kasama-twitch-integrations/internal/services"
	"github.com/blang/mpv"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gempir/go-twitch-irc/v4"
)

const mpvSocketPath = "/tmp/mpvsocket"
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

type mediaPlayer struct {
	cmd        *exec.Cmd
	player     *mpv.Client
	ipc        *mpv.IPCClient
	NowPlaying string
}

func newMediaPlayer() (*mediaPlayer, error) {
	cmd := exec.Command("mpv", "--no-video", "--idle=yes", "--input-ipc-server="+mpvSocketPath)
	if err := cmd.Start(); err != nil {
		logger.Errorf("Failed to start mpv: %s", err.Error())
		return nil, err
	}

	timeout := time.Now().Add(1 * time.Second)
	for {
		if time.Now().After(timeout) {
			return nil, fmt.Errorf("Timeout waiting for mpv to start")
		}
		_, err := net.Dial("unix", mpvSocketPath)
		if err == nil {
			break
		}
		logger.Debugf("Waiting for mpv to start")
		time.Sleep(50 * time.Millisecond)
	}

	ipc := mpv.NewIPCClient(mpvSocketPath)
	player := mpv.NewClient(ipc)

	return &mediaPlayer{
		cmd:        cmd,
		player:     player,
		ipc:        ipc,
		NowPlaying: "",
	}, nil
}

type AudioModule struct {
	done        chan bool
	stop        chan struct{}
	mediaPlayer *mediaPlayer
}

func NewAudioModule() *AudioModule {
	mediaPlayer, _ := newMediaPlayer()

	return &AudioModule{
		done:        nil,
		stop:        nil,
		mediaPlayer: mediaPlayer,
	}
}

func (m *AudioModule) Register() {
	events.Register(m.handleCommand)
	events.Register(m.handleStop)
	events.Register(m.handlePlayAudio)
	events.Register(m.handleDebugCommand)
	// events.Register(m.handleMediaPlayerEvent)

	err := speaker.Init(usualMP3SampleRate, usualMP3SampleRate.N(time.Second/10))
	if err != nil {
		logger.Errorf("Failed to start speaker with sample rate %fKHz: %s", float32(usualMP3SampleRate)/1000, err.Error())
	}
}

func (m *AudioModule) handleMediaPlayerEvent(event *mediaplayer.Event) error {
	switch event.Intent {
	case mediaplayer.MediaIntentPlay:
		paused, err := m.mediaPlayer.player.Pause()
		if err != nil {
			return err
		}
		err = m.mediaPlayer.player.SetPause(!paused)
		if err != nil {
			return err
		}
	case mediaplayer.MediaIntentNext:
		_, err := m.mediaPlayer.player.Exec("playlist-next", "force")
		if err != nil {
			return err
		}
	case mediaplayer.MediaIntentEnqueue:
		err := m.mediaPlayer.player.Loadfile(event.EnqueueQuery, mpv.LoadFileModeAppendPlay)
		return err
	}
	return nil
}

func (m *AudioModule) handleDebugCommand(message *twitch.PrivateMessage) error {
	if _, exists := message.User.Badges["broadcaster"]; !exists {
		return nil
	}

	if !strings.HasPrefix(message.Message, "!yt ") {
		return nil
	}
	video := strings.TrimPrefix(message.Message, "!yt ")

	if video == "what" {
		title, _ := m.mediaPlayer.player.GetProperty("media-title")
		logger.Debugf("Now playing: %s", title)
		return nil
	}

	err := m.mediaPlayer.player.Loadfile(video, mpv.LoadFileModeAppendPlay)
	return err
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
