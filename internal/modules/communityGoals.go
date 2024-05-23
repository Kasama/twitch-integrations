package modules

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/db"
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/gempir/go-twitch-irc/v4"
)

const coin_web_event_name = "community_coin"
const coin_timeout = 20 * time.Second

type CommunityGoalsModule struct {
	showingCoinUntil         time.Time
	currentGoalCommands      int
	currentGoalBeneficiaries map[string]*twitch.User
	db                       *db.Database
	twitchClient             *twitch.Client
	channel                  string
}

func NewCommunityGoalsModule(channel string) *CommunityGoalsModule {
	return &CommunityGoalsModule{
		showingCoinUntil:         time.Unix(0, 0),
		currentGoalCommands:      0,
		currentGoalBeneficiaries: make(map[string]*twitch.User),
		db:                       &db.Database{},
		twitchClient:             nil,
		channel:                  channel,
	}
}

type CommunityGoalCaptureEvent struct {
	beneficiaries []*twitch.User
}

func NewCommunityGoalCaptureEvent(beneficiaries []*twitch.User) *CommunityGoalCaptureEvent {
	return &CommunityGoalCaptureEvent{
		beneficiaries: beneficiaries,
	}
}

func (m *CommunityGoalsModule) setupDatabase() error {
	if m.db == nil {
		return fmt.Errorf("database is not initialized")
	}

	const createCommunityGoalsTable = `
		CREATE TABLE IF NOT EXISTS community_goals (
		goal_name TEXT PRIMARY KEY NOT NULL,
		needed_points INTEGER NOT NULL,
		active BOOLEAN,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`

	const createCommunityGoalsCollectTable = `
		CREATE TABLE IF NOT EXISTS community_goal_collects (
		username TEXT NOT NULL,
		userID TEXT NOT NULL,
		goal TEXT NOT NULL,
		count INTEGER DEFAULT 1,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (userID, goal));`

	_, err := m.db.Exec(createCommunityGoalsTable)
	if err != nil {
		return err
	}

	_, err = m.db.Exec(createCommunityGoalsCollectTable)
	if err != nil {
		return err
	}

	return nil
}

func (m *CommunityGoalsModule) collectGoal(userID, username string) error {
	goal, _, err := m.getActiveGoal()
	if err != nil {
		return err
	}
	if m.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	const insertGoal = `
		INSERT INTO community_goal_collects (userID, username, goal)
		VALUES (?, ?, ?)
		ON CONFLICT(userID, goal) DO UPDATE SET count=count+1, updated_at=CURRENT_TIMESTAMP;`
	_, err = m.db.Exec(insertGoal, userID, username, goal)
	if err != nil {
		return err
	}
	return nil
}

func (m *CommunityGoalsModule) setGoal(goal string, active bool) error {
	if m.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	const deactivate = `
		UPDATE community_goals SET active=false
		WHERE active=true;`
	_, err := m.db.Exec(deactivate, goal, active)
	if err != nil {
		return err
	}

	const insert = `
		INSERT INTO community_goals (goal_name, active)
		VALUES (?, ?)
		ON CONFLICT(goal_name) DO UPDATE SET active=excluded.active;`
	_, err = m.db.Exec(insert, goal, active)
	if err != nil {
		return err
	}
	return nil
}

func (m *CommunityGoalsModule) getActiveGoal() (string, int, error) {
	if m.db == nil {
		return "", 0, fmt.Errorf("database is not initialized")
	}
	const query = `
		SELECT goal_name, needed_points
		FROM community_goals
		WHERE active=true;`
	var goal string
	var goalPoints int
	err := m.db.QueryRow(query).Scan(&goal, &goalPoints)
	if err != nil {
		return "", 0, err
	}
	return goal, goalPoints, nil
}

func (m *CommunityGoalsModule) getGoalCollects(goal string) (int, error) {
	if m.db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	const query = `
		SELECT count
		FROM community_goal_collects
		WHERE goal=?;`

	rows, err := m.db.Query(query, goal)
	if err != nil {
		return 0, err
	}

	count := 0
	for {
		if !rows.Next() {
			break
		}
		var c int
		err = rows.Scan(&c)
		if err != nil {
			break
		}
		count += c
	}

	return count, nil
}

func (m *CommunityGoalsModule) Register() {
	db, err := db.GetDatabase()
	if err != nil {
		logger.Errorf("Failed to open database: %s", err.Error())
	}
	m.db = db
	err = m.setupDatabase()
	if err != nil {
		logger.Errorf("Failed to setup database: %s", err.Error())
	}

	events.Register(m.handleCommand)
	events.Register(m.handleTicker)
	events.Register(m.handleCaptureEvent)
	events.Register(m.handleTwitchClient)
	events.Register(m.handleHelp)
}

func clearCoin() {
	events.Dispatch(NewWebEvent(coin_web_event_name, ""))
}

func showCoin(x, y int) {
	events.Dispatch(NewWebEvent(coin_web_event_name, views.RenderToString(views.CommunityCoin(x, y))))
}

func (m *CommunityGoalsModule) handleTwitchClient(client *twitch.Client) error {
	m.twitchClient = client
	return nil
}

func (m *CommunityGoalsModule) handleTicker(t *time.Time) error {
	n := rand.Int31n(600)
	if n != 0 || time.Now().Before(m.showingCoinUntil) {
		return nil
	}

	m.currentGoalCommands = 0
	m.currentGoalBeneficiaries = map[string]*twitch.User{}
	x := rand.Intn(90) + 5
	y := rand.Intn(90) + 5
	showCoin(x, y)
	m.showingCoinUntil = time.Now().Add(coin_timeout)

	go func() {
		time.Sleep(coin_timeout)
		clearCoin()
		m.currentGoalCommands = 0
		m.currentGoalBeneficiaries = map[string]*twitch.User{}
	}()

	logger.Debug("Showing coin")

	return nil
}

func (m *CommunityGoalsModule) handleCaptureEvent(event *CommunityGoalCaptureEvent) error {

	for _, b := range event.beneficiaries {
		err := m.collectGoal(b.ID, b.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *CommunityGoalsModule) handleHelp(msg *twitch.PrivateMessage) error {
	if !strings.HasPrefix(msg.Message, "!objetivo") {
		return nil
	}

	currentGoal, neededGoalPoints, err := m.getActiveGoal()
	if err != nil {
		return err
	}

	goalPoints, err := m.getGoalCollects(currentGoal)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("O objetivo atual é: %s. Atualmente temos %d/%d pontos necessários. Use !coletar na hora certa para contribuir", currentGoal, goalPoints, neededGoalPoints)

	if m.twitchClient != nil {
		logger.Debugf("id: %s, Got client: %v", m.channel, m.twitchClient)
		m.twitchClient.Say(m.channel, text)
	} else {
		logger.Debugf("Twitch client was nil, but was gonna say: '%s'", text)
	}

	return nil
}

func (m *CommunityGoalsModule) handleCommand(msg *twitch.PrivateMessage) error {
	if !strings.HasPrefix(msg.Message, "!coletar") {
		return nil
	}

	if time.Now().After(m.showingCoinUntil) {
		return nil
	}
	if _, exists := m.currentGoalBeneficiaries[msg.User.ID]; !exists {
		m.currentGoalCommands += 1
		m.currentGoalBeneficiaries[msg.User.ID] = &msg.User
		logger.Debugf("Collected coin, count: %d", m.currentGoalCommands)
	}

	if m.currentGoalCommands >= 2 {
		m.currentGoalCommands = 0
		beneficiaries := make([]*twitch.User, 0, len(m.currentGoalBeneficiaries))
		for _, b := range m.currentGoalBeneficiaries {
			beneficiaries = append(beneficiaries, b)
		}
		logger.Debugf("Collected coin, count: %d, beneficiaries: %v, map: %v", len(beneficiaries), beneficiaries, m.currentGoalBeneficiaries)
		m.currentGoalBeneficiaries = make(map[string]*twitch.User)
		m.showingCoinUntil = time.Unix(0, 0)
		clearCoin()
		events.Dispatch(NewCommunityGoalCaptureEvent(beneficiaries))
	}

	return nil
}

var _ events.EventHandler = &CommunityGoalsModule{}
