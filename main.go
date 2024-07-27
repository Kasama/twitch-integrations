package main

import (
	"context"
	"log"
	"os"

	"github.com/Kasama/kasama-twitch-integrations/internal/http"
	"github.com/Kasama/kasama-twitch-integrations/internal/modules"
	"github.com/Kasama/kasama-twitch-integrations/internal/services"
	"github.com/Kasama/kasama-twitch-integrations/internal/spotify"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // ignore errors

	logger := log.Default()

	twitchClientId, exists := os.LookupEnv("TWITCH_CLIENT_ID")
	if !exists {
		logger.Fatal("var TWITCH_CLIENT_ID not found")
	}

	twitchClientSecret, exists := os.LookupEnv("TWITCH_CLIENT_SECRET")
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

	spotifyClientId, exists := os.LookupEnv("SPOTIFY_CLIENT_ID")
	if !exists {
		logger.Fatal("var SPOTIFY_CLIENT_ID not found")
	}

	spotifyClientSecret, exists := os.LookupEnv("SPOTIFY_CLIENT_SECRET")
	if !exists {
		logger.Fatal("var SPOTIFY_CLIENT_SECRET not found")
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

	twitchConfig := twitch.NewTwitchConfig(twitchClientId, twitchClientSecret, twitchUserId, twitchUsername, "http://localhost:3000/auth/twitch/redirect")
	spotifyConfig := spotify.NewSpotifyConfig(spotifyClientId, spotifyClientSecret, "http://localhost:3000/auth/spotify/redirect")
	webEventsModule := modules.NewWebEventsModule()
	spotifyModule := modules.NewSpotifyModule(appContext, twitchUsername)

	// Register modules
	modules.NewYappingModule().Register()
	modules.NewDiceModule().Register()
	modules.NewCalabocaModule(twitchUserId).Register()
	modules.NewAudioModule().Register()
	modules.NewTwitchHelixModule().Register()
	modules.NewTimeoutModule(twitchUserId).Register()
	modules.NewUserThemeModule(twitchUsername).Register()
	modules.NewCommunityGoalsModule().Register()
	spotifyModule.Register()
	modules.NewCarteiradaModule(twitchUsername).Register()
	modules.NewGringouModule().Register()
	modules.NewPortugouModule().Register()
	modules.NewEmouModule().Register()
	modules.NewMacropadModule().Register()
	modules.NewCountersModule(twitchUsername).Register()
	modules.NewChopanModule().Register()
	modules.NewFloresbertoModule().Register()
	modules.NewCommandsModule().Register()
	modules.NewMediaPlayerModule().Register()
	modules.NewBotBansModule(twitchUserId).Register()
	webEventsModule.Register()

	// Register services
	services.NewTwitchChatService(twitchUsername).Register()
	services.NewTwitchEventSubService(twitchUserId).Register()
	services.NewOBSService(obsAddress, obsPassword).Register()
	services.NewTimerService(appContext).Register()
	services.NewSpotifyService(appContext, spotifyConfig).Register()
	services.NewOmegaStrikersService(appContext).Register()

	// Start server
	_ = http.NewHandlers(environment, twitchConfig, spotifyConfig, webEventsModule, spotifyModule.Queue()).Start("0.0.0.0", "3000")
}
