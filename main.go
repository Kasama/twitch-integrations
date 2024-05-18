package main

import (
	"context"
	"log"
	"os"

	"github.com/Kasama/kasama-twitch-integrations/internal/http"
	"github.com/Kasama/kasama-twitch-integrations/internal/modules"
	"github.com/Kasama/kasama-twitch-integrations/internal/services"
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

	obsAddress, exists := os.LookupEnv("OBS_WS_URL")
	if !exists {
		obsAddress = "localhost:4455"
	}

	obsPassword, exists := os.LookupEnv("OBS_WS_PASSWORD")
	if !exists {
		logger.Fatal("var OBS_WS_PASSWORD not found")
	}

	t, exists := os.LookupEnv("STREAM_TEMP_DIR")
	if !exists {
		t = "twitch-streaming"
	}
	tmpDir, err := os.MkdirTemp("", t)
	if err != nil {
		logger.Fatalf("Could not create tempdir %s", tmpDir)
	}

	appContext := context.Background()

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
			"moderation:read",
			"moderator:manage:banned_users",
			"moderator:manage:chat_messages",
			"channel:read:redemptions",
			"channel:manage:redemptions",
			"channel:read:predictions",
			"channel:manage:predictions",
		},
	}
	twitchConfig := twitch.NewTwitchConfig(clientId, clientSecret, twitchUserId, twitchUsername, oauth2Config)
	webEventsModule := modules.NewWebEventsModule()

	// Register modules
	modules.NewYappingModule(twitchUsername).Register()
	modules.NewDiceModule().Register()
	modules.NewCalabocaModule(twitchUserId).Register()
	modules.NewAudioModule().Register()
	modules.NewTwitchHelixModule().Register()
	modules.NewTimeoutModule(twitchUserId).Register()
	modules.NewUserThemeModule(twitchUsername).Register()
	modules.NewCommunityGoalsModule().Register()
	webEventsModule.Register()
	// modules.NewSpotifyModule(spotifyConfig.clientId, spotifyConfig.clientSecret).Register()

	// Register services
	services.NewTwitchChatService(twitchUsername).Register()
	services.NewTwitchEventSubService(twitchUserId).Register()
	services.NewOBSService(obsAddress, obsPassword).Register()
	services.NewTimerService(appContext).Register()

	// Start server
	_ = http.NewHandlers(environment, twitchConfig, webEventsModule).Start("0.0.0.0", "3000")
}
