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
	TwitchConfig *TwitchConfig
}

func NewTwitchAuth(token *oauth2.Token, config *TwitchConfig) *TwitchAuth {
	return &TwitchAuth{
		Token:        token,
		TwitchConfig: config,
	}
}

type TwitchConfig struct {
	ClientId     string
	ClientSecret string
	Oauth2config *oauth2.Config
	UserID       string
	Channel      string
}

func NewTwitchConfig(clientId, clientSecret, userId, channel, redirectURL string) *TwitchConfig {
	oauth2Config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     twitchAuth.Endpoint,
		RedirectURL:  redirectURL,
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
			"moderator:manage:automod",
			"channel:read:redemptions",
			"channel:manage:redemptions",
			"channel:read:predictions",
			"channel:manage:predictions",
		},
	}

	return &TwitchConfig{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Oauth2config: oauth2Config,
		UserID:       userId,
		Channel:      channel,
	}
}

func (t *TwitchAuth) RefreshAuthToken() (*TwitchAuth, error) {
	if t.Token == nil {
		return t, nil
	}

	resp, err := http.Post(twitchAuth.Endpoint.TokenURL, "application/x-www-form-urlencoded",
		strings.NewReader("grant_type=refresh_token&refresh_token="+t.RefreshToken+"&client_id="+t.TwitchConfig.ClientId+"&client_secret="+t.TwitchConfig.ClientSecret),
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
		Token:        newToken,
		TwitchConfig: t.TwitchConfig,
	}, nil
}
