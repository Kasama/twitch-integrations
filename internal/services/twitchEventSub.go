package services

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	internalTwitch "github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

type TwitchEventSubService struct {
	userID string
	auth   *internalTwitch.TwitchAuth
	exit   chan struct{}
}

func NewTwitchEventSubService(userID string) *TwitchEventSubService {
	return &TwitchEventSubService{
		userID: userID,
	}
}

func (t *TwitchEventSubService) Register() {
	events.Register(t.handleAuth)
}

func (t *TwitchEventSubService) handleAuth(auth *internalTwitch.TwitchAuth) error {
	if t.exit != nil {
		close(t.exit)
	}
	exit := make(chan struct{})
	t.exit = exit

	t.auth = auth

	client := twitch.NewClient()

	client.OnError(func(err error) {
		logger.Errorf("TwitchEventSub client error: %v", err)
	})
	client.OnWelcome(func(message twitch.WelcomeMessage) {
		logger.Debugf("WELCOME: %v\n", message)

		events := []twitch.EventSubscription{
			twitch.SubStreamOnline,
			twitch.SubChannelChannelPointsCustomRewardRedemptionAdd,
			twitch.SubChannelChannelPointsCustomRewardRedemptionUpdate,
			twitch.SubChannelChannelPointsCustomRewardRemove,
			twitch.SubStreamOffline,
		}

		for _, event := range events {
			logger.Printf("subscribing to %s\n", event)
			_, err := twitch.SubscribeEvent(twitch.SubscribeRequest{
				SessionID:   message.Payload.Session.ID,
				ClientID:    auth.TwitchConfig.ClientId,
				AccessToken: auth.AccessToken,
				Event:       event,
				Condition: map[string]string{
					"broadcaster_user_id": t.userID,
				},
			})
			if err != nil {
				logger.Errorf("twitchEventSub failed to subcribe to events: %v\n", err)
				return
			}
		}
	})
	client.OnEventChannelChannelPointsCustomRewardRedemptionAdd(func(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
		events.Dispatch(&event)
	})
	client.OnRevoke(func(message twitch.RevokeMessage) {
		logger.Printf("REVOKE: %v\n", message)
	})

	go func() {
		<-exit
		_ = client.Close()
	}()

	go func() {
		err := client.Connect()
		if err != nil {
			logger.Errorf("failed to connect to twitch eventsub: %v\n", err)
		}
	}()

	return nil
}
