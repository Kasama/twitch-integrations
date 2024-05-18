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

const twitchAuthToken = "twitch-auth-token"
const twitchRefreshToken = "twitch-refresh-token"
const twitchTokenExpiration = "twitch-token-expiration"
const commonTimeFormat = time.RFC3339Nano

func TwitchAuthFromCookies(c echo.Context, twitchConfig *twitch.TwitchConfig) (*twitch.TwitchAuth, error) {
	authToken, err := c.Cookie(twitchAuthToken)
	if err != nil {
		return nil, err
	}

	refreshToken, err := c.Cookie(twitchRefreshToken)
	if err != nil {
		return nil, err
	}

	expiry, err := c.Cookie(twitchTokenExpiration)
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
	SaveToCookies(c, auth)

	events.Dispatch[*twitch.TwitchAuth](auth)

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (t *TwitchHandler) handleIndex(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("Twitch configs", views.TwitchIndex()))
}
