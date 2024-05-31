package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const commonTimeFormat = time.RFC3339Nano

func TwitchAuthFromCookies(c echo.Context, twitchConfig *twitch.TwitchConfig) (*twitch.TwitchAuth, error) {
	authToken, err := c.Cookie(twitchCookiePrefix + "-" + authTokenCookieName)
	if err != nil {
		return nil, err
	}

	refreshToken, err := c.Cookie(twitchCookiePrefix + "-" + refreshTokenCookieName)
	if err != nil {
		return nil, err
	}

	expiry, err := c.Cookie(twitchCookiePrefix + "-" + tokenExpirationCookieName)
	if err != nil {
		return nil, err
	}
	expiryTime, _ := time.Parse(commonTimeFormat, expiry.Value)

	return twitch.NewTwitchAuth(
		&oauth2.Token{
			AccessToken:  authToken.Value,
			RefreshToken: refreshToken.Value,
			Expiry:       expiryTime,
		},
		twitchConfig,
	), nil
}

type TwitchHandler struct {
	twitchConfig *twitch.TwitchConfig
}

func NewTwitchHandler(twitchConfig *twitch.TwitchConfig) *TwitchHandler {
	return &TwitchHandler{
		twitchConfig: twitchConfig,
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

	auth := twitch.NewTwitchAuth(token, t.twitchConfig)
	SaveTokenToCookies(c, twitchCookiePrefix, auth.Token)

	events.Dispatch(auth)

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (t *TwitchHandler) handleIndex(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("Twitch configs", views.TwitchIndex()))
}
