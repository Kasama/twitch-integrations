package twitch

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/joeyak/go-twitch-eventsub/v2"
)

func SetupEventSub(clientID, accessToken, userID string, dispatcher *events.EventDispatcher) {
	l := dispatcher.Context.Logger
	client := twitch.NewClient()

	client.OnError(func(err error) {
		l.Printf("ERROR: %v\n", err)
	})
	client.OnWelcome(func(message twitch.WelcomeMessage) {
		l.Printf("WELCOME: %v\n", message)

		events := []twitch.EventSubscription{
			twitch.SubStreamOnline,
			twitch.SubChannelChannelPointsCustomRewardRedemptionAdd,
			twitch.SubChannelChannelPointsCustomRewardRedemptionUpdate,
			twitch.SubChannelChannelPointsCustomRewardRemove,
			twitch.SubStreamOffline,
		}

		for _, event := range events {
			l.Printf("subscribing to %s\n", event)
			_, err := twitch.SubscribeEvent(twitch.SubscribeRequest{
				SessionID:   message.Payload.Session.ID,
				ClientID:    clientID,
				AccessToken: accessToken,
				Event:       event,
				Condition: map[string]string{
					"broadcaster_user_id": userID,
				},
			})
			if err != nil {
				l.Printf("ERROR: %v\n", err)
				return
			}
		}
	})
	client.OnNotification(func(message twitch.NotificationMessage) {
		l.Printf("NOTIFICATION: %s: %#v\n", message.Payload.Subscription.Type, message.Payload.Event)
	})
	client.OnEventChannelChannelPointsCustomRewardRedemptionAdd(func(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
		dispatcher.Dispatch(&events.Event{
			Kind:                          events.EventKindChannelPointsRewardRedemption,
			ChannelPointsRewardRedemption: &event,
		})
	})
	client.OnRevoke(func(message twitch.RevokeMessage) {
		l.Printf("REVOKE: %v\n", message)
	})

	err := client.Connect()
	if err != nil {
		l.Printf("Could not connect client: %v\n", err)
	}
}
