package global

import (
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"golang.org/x/oauth2"
)

type GlobalType struct {
	twitchAuthConfig *twitch.TwitchConfig
	twitchAuth       *twitch.TwitchAuth
}

var Global *GlobalType = &GlobalType{}

func (g *GlobalType) SetTwitchAuth(auth *twitch.TwitchAuth) {
	g.twitchAuth = auth
}

func (g *GlobalType) GetTwitchToken() *oauth2.Token {
	if g.twitchAuth == nil {
		return nil
	}

	if g.twitchAuth.Expiry.Add(-10 * time.Minute).Before(time.Now()) {
		newToken, err := g.twitchAuthConfig.RefreshTwitchAuthToken(g.twitchAuth)
		if err != nil {
			return nil
		}
		g.twitchAuth = newToken
	}

	return g.twitchAuth.Token
}
