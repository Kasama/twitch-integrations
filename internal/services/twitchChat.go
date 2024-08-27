package services

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	twitchChat "github.com/gempir/go-twitch-irc/v4"
)

type TwitchChatService struct {
	channel    string
	twitchAuth *twitch.TwitchAuth
	exit       chan struct{}
}

// Register implements events.EventHandler.
func (t *TwitchChatService) Register() {
	events.Register(t.handleToken)
}

func NewTwitchChatService(channel string) *TwitchChatService {
	return &TwitchChatService{
		channel:    channel,
		twitchAuth: nil,
		exit:       nil,
	}
}

type EventConnected struct{}

func (t *TwitchChatService) handleToken(token *twitch.TwitchAuth) error {
	if t.exit != nil {
		close(t.exit)
	}

	t.twitchAuth = token

	exit := make(chan struct{})
	t.exit = exit

	client := twitchChat.NewClient(t.channel, "oauth:"+token.AccessToken)

	client.OnPrivateMessage(func(message twitchChat.PrivateMessage) {
		logger.Debugf("@%s: %s", message.User.DisplayName, message.Message)
		events.Dispatch(&message)
	})

	client.OnConnect(func() {
		events.Dispatch(EventConnected{})
		logger.Debug("Chat Connected")
	})

	client.OnUserNoticeMessage(func(message twitchChat.UserNoticeMessage) {
		events.Dispatch(&message)
	})

	client.Capabilities = []string{twitchChat.TagsCapability, twitchChat.MembershipCapability, twitchChat.CommandsCapability}

	logger.Debug("Joining chat")
	client.Join(t.channel)

	go func() {
		<-exit
		logger.Debug("Got chat exit channel message")
		_ = client.Disconnect()
	}()

	go func() {
		logger.Debug("Connecting chat")
		err := client.Connect()
		if err != nil {
			logger.Errorf("failed to connect to chat: %v", err)
		}
		logger.Debug("Disconnecting from chat")
	}()

	events.Dispatch(client)

	return nil
}

var _ events.EventHandler = &TwitchChatService{}
