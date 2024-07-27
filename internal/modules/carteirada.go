package modules

import (
	"fmt"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/gempir/go-twitch-irc/v4"
)

type CarteiradaModule struct {
	channel          string
}

func NewCarteiradaModule(channel string) *CarteiradaModule {
	return &CarteiradaModule{
		channel:          channel,
	}
}

// Register implements events.EventHandler.
func (m *CarteiradaModule) Register() {
	events.Register(m.handleCommand)
}

func (m *CarteiradaModule) handleCommand(msg *twitch.PrivateMessage) error {
	if !strings.Contains(strings.TrimSpace(msg.Message), "!carteirada") {
		return nil
	}

	var target string
	parts := strings.Split(strings.TrimSpace(msg.Message), " ")
	if msg.Reply != nil {
		target = msg.Reply.ParentDisplayName
	} else {
		if len(parts) < 2 {
			target = msg.User.DisplayName
		} else {
			target = strings.Trim(parts[1], "@")
		}
	}

	events.Dispatch(NewChatMessage(fmt.Sprintf("Ah é? Mas quantos títulos vc tem? @%s", target)))

	return nil
}

var _ events.EventHandler = &CarteiradaModule{}
