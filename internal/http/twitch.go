package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/global"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/services"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const twitchAuthToken = "twitch-auth-token"
const twitchRefreshToken = "twitch-refresh-token"
const twitchTokenExpiration = "twitch-token-expiration"
const commonTimeFormat = time.RFC3339Nano

func TwitchAuthFromCookies(c echo.Context) (*twitch.TwitchAuth, error) {
	authToken, err := c.Cookie(twitchAuthToken)
	if err != nil {
		fmt.Printf("Error getting auth cookie: %v\n", err)
		return nil, err
	}
	refreshToken, err := c.Cookie(twitchRefreshToken)
	if err != nil {
		fmt.Printf("Error getting refresh cookie: %v\n", err)
		return nil, err
	}
	expiry, err := c.Cookie(twitchTokenExpiration)
	if err != nil {
		fmt.Printf("Error getting expiry cookie: %v\n", err)
		return nil, err
	}
	expiryTime, _ := time.Parse(commonTimeFormat, expiry.Value)
	return &twitch.TwitchAuth{
		Token: &oauth2.Token{
			AccessToken:  authToken.Value,
			RefreshToken: refreshToken.Value,
			Expiry:       expiryTime,
		},
	}, nil
}

func SaveToCookies(c echo.Context, t *twitch.TwitchAuth) {
	c.SetCookie(&http.Cookie{
		Name:  twitchAuthToken,
		Value: t.AccessToken,
		Path:  "/",
	})
	c.SetCookie(&http.Cookie{
		Name:  twitchRefreshToken,
		Value: t.RefreshToken,
		Path:  "/",
	})
	c.SetCookie(&http.Cookie{
		Name:  twitchTokenExpiration,
		Value: t.Expiry.Format(commonTimeFormat),
		Path:  "/",
	})
}

type TwitchHandler struct {
	twitchConfig     *twitch.TwitchConfig
	eventsDispatcher *events.EventDispatcher
	chatService      *services.TwitchChatService
	chatServiceExit  chan struct{}
}

func NewTwitchHandler(twitchConfig *twitch.TwitchConfig, eventsDispatcher *events.EventDispatcher) *TwitchHandler {
	return &TwitchHandler{
		twitchConfig:     twitchConfig,
		eventsDispatcher: eventsDispatcher,
	}
}

func (t *TwitchHandler) handleAuth(c echo.Context) error {
	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return err
	}

	state := hex.EncodeToString(tokenBytes[:])

	url := t.twitchConfig.Oauth2config.AuthCodeURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (t *TwitchHandler) handleRedirect(c echo.Context) error {
	token, err := t.twitchConfig.Oauth2config.Exchange(context.Background(), c.QueryParam("code"))
	if err != nil {
		return err
	}

	auth := &twitch.TwitchAuth{Token: token}
	SaveToCookies(c, auth)

	global.Global.SetTwitchAuth(auth)

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (t *TwitchHandler) handleEnableChatService(c echo.Context) error {

	if t.chatService != nil {
		return Render(c, http.StatusBadRequest, views.Flash("Chat service already enabled"))
	}

	service := services.NewTwitchChatService(t.twitchConfig.Channel)
	exit := service.Start(t.eventsDispatcher)
	if exit == nil {
		return Render(c, http.StatusInternalServerError, views.Flash("Chat service could not be started"))
	}

	t.chatService = service
	t.chatServiceExit = exit

	return c.NoContent(http.StatusOK)
}

func (t *TwitchHandler) handleDisableChatService(c echo.Context) error {

	if t.chatService == nil {
		return Render(c, http.StatusBadRequest, views.Flash("Chat service is not enabled"))
	}

	t.chatService.Stop(t.chatServiceExit)
	t.chatService = nil
	t.chatServiceExit = nil

	return c.NoContent(http.StatusOK)
}

func (t *TwitchHandler) handleIndex(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("Twitch configs", views.TwitchIndex()))
}
