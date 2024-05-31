package spotify

import (
	spotifyAuth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type SpotifyAuth struct {
	*oauth2.Token
	SpotifyConfig *SpotifyConfig
}

func NewSpotifyAuth(token *oauth2.Token, config *SpotifyConfig) *SpotifyAuth {
	return &SpotifyAuth{
		Token:         token,
		SpotifyConfig: config,
	}
}

type SpotifyConfig struct {
	Oauth2config *oauth2.Config
}

func NewSpotifyConfig(clientId, clientSecret, redirectURL string) *SpotifyConfig {
	return &SpotifyConfig{
		Oauth2config: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   spotifyAuth.AuthURL,
				TokenURL:  spotifyAuth.TokenURL,
				AuthStyle: oauth2.AuthStyleInParams,
			},
			RedirectURL: redirectURL,
			Scopes:      []string{spotifyAuth.ScopeStreaming, spotifyAuth.ScopePlaylistModifyPrivate, spotifyAuth.ScopeUserModifyPlaybackState, spotifyAuth.ScopeUserReadPlaybackState},
		},
	}
}
