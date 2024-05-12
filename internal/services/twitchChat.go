package services

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/global"
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

// Start implements events.EventEmitter.
func (t *TwitchChatService) Start(dispatcher *events.EventDispatcher) chan struct{} {
	token := global.Global.GetTwitchToken()
	if token == nil {
		return nil
	}
	exit := make(chan struct{})

	l := dispatcher.Context.Logger
	client := twitch.NewClient(t.channel, "oauth:"+token.AccessToken)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		dispatcher.Dispatch(&events.Event{
			Kind:        events.EventKindChatMessage,
			ChatMessage: &message,
		})
	})

	client.OnConnect(func() {
		dispatcher.Dispatch(&events.Event{
			Kind: events.EventKindConnected,
		})
		l.Println("Chat Connected")
	})

	client.Capabilities = []string{twitch.TagsCapability, twitch.MembershipCapability, twitch.CommandsCapability}

	l.Println("Joining chat")
	client.Join(t.channel)

	go func() {
		<-exit
		_ = client.Disconnect()
	}()

	go func() {
		l.Println("Connecting chat")
		err := client.Connect()
		if err != nil {
			l.Fatal(err)
		}
	}()

	dispatcher.Dispatch(&events.Event{
		Kind:                          events.EventKindChatClientCreated,
		ChatClientCreated:             client,
	})

	return exit
}

// Stop implements events.EventEmitter.
func (*TwitchChatService) Stop(exit chan struct{}) {
	close(exit)
}

var _ events.EventEmitter = &TwitchChatService{}
