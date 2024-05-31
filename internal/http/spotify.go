package http

import (
	"context"
	"net/http"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/spotify"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const state = "spotify-twitch-auth"

type SpotifyHandler struct {
	spotifyConfig *spotify.SpotifyConfig
}

func NewSpotifyHandler(cfg *spotify.SpotifyConfig) *SpotifyHandler {
	return &SpotifyHandler{
		spotifyConfig: cfg,
	}
}

func SpotifyAuthFromCookies(c echo.Context, spotifyConfig *spotify.SpotifyConfig) (*spotify.SpotifyAuth, error) {
	authToken, err := c.Cookie(spotifyCookiePrefix + "-" + authTokenCookieName)
	if err != nil {
		return nil, err
	}

	refreshToken, err := c.Cookie(spotifyCookiePrefix + "-" + refreshTokenCookieName)
	if err != nil {
		return nil, err
	}

	expiry, err := c.Cookie(spotifyCookiePrefix + "-" + tokenExpirationCookieName)
	if err != nil {
		return nil, err
	}
	expiryTime, _ := time.Parse(commonTimeFormat, expiry.Value)

	return spotify.NewSpotifyAuth(
		&oauth2.Token{
			AccessToken:  authToken.Value,
			RefreshToken: refreshToken.Value,
			Expiry:       expiryTime,
		},
		spotifyConfig,
	), nil
}

func (s *SpotifyHandler) handleAuth(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, s.spotifyConfig.Oauth2config.AuthCodeURL(state))
}

func (s *SpotifyHandler) handleRedirect(c echo.Context) error {
	token, err := s.spotifyConfig.Oauth2config.Exchange(context.Background(), c.QueryParam("code"))
	if err != nil {
		return err
	}

	auth := spotify.NewSpotifyAuth(token, s.spotifyConfig)
	SaveTokenToCookies(c, spotifyCookiePrefix, auth.Token)

	events.Dispatch(auth)

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
