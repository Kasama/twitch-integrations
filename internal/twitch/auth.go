package twitch

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	twitchAuth "golang.org/x/oauth2/twitch"
)

type TwitchAuth struct {
	*oauth2.Token
}

type TwitchConfig struct {
	ClientId     string
	ClientSecret string
	Oauth2config *oauth2.Config
	UserID       string
	Channel      string
}

func NewTwitchConfig(clientId, clientSecret, userId, channel string, oauth2config *oauth2.Config) *TwitchConfig {
	return &TwitchConfig{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Oauth2config: oauth2config,
		UserID:       userId,
		Channel:      channel,
	}
}

func (t *TwitchConfig) RefreshTwitchAuthToken(token *TwitchAuth) (*TwitchAuth, error) {

	resp, err := http.Post(twitchAuth.Endpoint.TokenURL, "application/x-www-form-urlencoded",
		strings.NewReader("grant_type=refresh_token&refresh_token="+token.RefreshToken+"&client_id="+t.ClientId+"&client_secret="+t.ClientSecret),
	)
	if err != nil {
		return nil, err
	}

	responseBody := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(responseBody)
	if err != nil {
		return nil, err
	}

	var newToken *oauth2.Token
	err = json.Unmarshal(responseBody, &newToken)
	if err != nil {
		return nil, err
	}

	return &TwitchAuth{
		Token: newToken,
	}, nil

}
