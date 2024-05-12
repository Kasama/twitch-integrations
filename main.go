package main

import (
	"log"
	"os"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http"
	"github.com/Kasama/kasama-twitch-integrations/internal/modules"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	twitchAuth "golang.org/x/oauth2/twitch"
)

func main() {
	_ = godotenv.Load() // ignore errors

	logger := log.Default()

	clientId, exists := os.LookupEnv("TWITCH_CLIENT_ID")
	if !exists {
		logger.Fatal("var TWITCH_CLIENT_ID not found")
	}

	clientSecret, exists := os.LookupEnv("TWITCH_CLIENT_SECRET")
	if !exists {
		logger.Fatal("var TWITCH_CLIENT_SECRET not found")
	}

	twitchUsername, exists := os.LookupEnv("TWITCH_USERNAME")
	if !exists {
		logger.Fatal("var TWITCH_USERNAME not found")
	}

	twitchUserId, exists := os.LookupEnv("TWITCH_USERID")
	if !exists {
		logger.Fatal("var TWITCH_USERID not found")
	}

	environment, exists := os.LookupEnv("ENVIRONMENT")
	if !exists {
		environment = "development"
	}

	dispatcher := events.NewEventDispatcher(events.WithLogger(log.Default()))
	dispatcher.RegisterHandler(modules.NewDiceModule())
	dispatcherChannel := dispatcher.StartAsync()

	oauth2Config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     twitchAuth.Endpoint,
		RedirectURL:  "http://localhost:3000/auth/twitch/redirect",
		// list of scopes at https://dev.twitch.tv/docs/authentication/scopes/
		Scopes: []string{
			"user:read:email",
			"chat:read",
			"chat:edit",
			"channel:bot",
			"channel:moderate",
			"user:bot",
			"user:read:chat",
			"user:write:chat",
			"whispers:read",
			"whispers:edit",
			"channel:manage:redemptions",
			"channel:manage:polls",
		},
	}
	twitchConfig := twitch.NewTwitchConfig(clientId, clientSecret, twitchUserId, twitchUsername, oauth2Config)

	// Register modules
	dispatcher.RegisterHandler(modules.NewYappingModule(twitchUsername))

	// Start server
	_ = http.NewHandlers(environment, twitchConfig, dispatcher).Start("localhost", "3000")

	<-dispatcherChannel
}
