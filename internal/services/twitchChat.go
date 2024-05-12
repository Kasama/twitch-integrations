package services

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/global"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchChatService struct {
	channel string
}

func NewTwitchChatService(channel string) *TwitchChatService {
	return &TwitchChatService{
		channel,
	}
}

type EventConnected struct{}

// Start implements events.EventEmitter.
func (t *TwitchChatService) Start() chan struct{} {
	token := global.Global.GetTwitchToken()
	if token == nil {
		return nil
	}
	exit := make(chan struct{})

	client := twitch.NewClient(t.channel, "oauth:"+token.AccessToken)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		events.Dispatch(&message)
	})

	client.OnConnect(func() {
		events.Dispatch(EventConnected{})
		logger.Debug("Chat Connected")
	})

	client.Capabilities = []string{twitch.TagsCapability, twitch.MembershipCapability, twitch.CommandsCapability}

	logger.Debug("Joining chat")
	client.Join(t.channel)

	go func() {
		<-exit
		_ = client.Disconnect()
	}()

	go func() {
		logger.Debug("Connecting chat")
		err := client.Connect()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	events.Dispatch(client)

	return exit
}

// Stop implements events.EventEmitter.
func (*TwitchChatService) Stop(exit chan struct{}) {
	close(exit)
}

var _ events.EventEmitter = &TwitchChatService{}
