package modules

import (
	"math/rand"
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/gempir/go-twitch-irc/v4"
)

const coin_web_event_name = "community_coin"
const coin_timeout = 10 * time.Second

type CommunityGoalsModule struct {
	showingCoinUntil    time.Time
	currentGoalCommands int
}

func NewCommunityGoalsModule() *CommunityGoalsModule {
	return &CommunityGoalsModule{
		showingCoinUntil:    time.Unix(0, 0),
		currentGoalCommands: 0,
	}
}

type CommunityGoalCaptureEvent struct{}

func NewCommunityGoalCaptureEvent() *CommunityGoalCaptureEvent {
	return &CommunityGoalCaptureEvent{}
}

func (m *CommunityGoalsModule) Register() {
	events.Register(m.handleCommand)
	events.Register(m.handleTicker)
}

func clearCoin() {
		events.Dispatch(NewWebEvent(coin_web_event_name, ""))
}

func showCoin(x, y int) {
	events.Dispatch(NewWebEvent(coin_web_event_name, views.RenderToString(views.CommunityCoin(x, y))))
}

func (m *CommunityGoalsModule) handleTicker(t *time.Time) error {
	n := rand.Int31n(10)
	if n != 0 || time.Now().Before(m.showingCoinUntil) {
		return nil
	}

	m.currentGoalCommands = 0
	x := rand.Intn(90) + 5
	y := rand.Intn(90) + 5
	showCoin(x, y)
	m.showingCoinUntil = time.Now().Add(coin_timeout)

	go func() {
		time.Sleep(coin_timeout)
		clearCoin()
		m.currentGoalCommands = 0
	}()

	logger.Debug("Showing coin")

	return nil
}

func (m *CommunityGoalsModule) handleCommand(msg *twitch.PrivateMessage) error {
	if !strings.HasPrefix(msg.Message, "!coletar") {
		return nil
	}

	if time.Now().After(m.showingCoinUntil) {
		return nil
	}
	m.currentGoalCommands += 1
	logger.Debugf("Collected coin, count: %d", m.currentGoalCommands)

	if m.currentGoalCommands >= 2 {
		m.currentGoalCommands = 0
		m.showingCoinUntil = time.Unix(0, 0)
		events.Dispatch(NewCommunityGoalCaptureEvent())
		clearCoin()
	}

	return nil
}

var _ events.EventHandler = &CommunityGoalsModule{}
