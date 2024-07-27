package modules

import (
	"fmt"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/db"
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/gempir/go-twitch-irc/v4"
)

type CounterName string

const (
	KasamadasCount CounterName = "kasamadas"
)

func intoCounterName(name string) (CounterName, error) {
	switch name {
	case string(KasamadasCount):
		return KasamadasCount, nil
	}

	return "", fmt.Errorf("invalid counter name: %s", name)
}

type CountersModule struct {
	db               *db.Database
	channel          string
}

func NewCountersModule(channel string) *CountersModule {
	return &CountersModule{
		db:               nil,
		channel:          channel,
	}
}

func (m *CountersModule) sendTwitchChatMessage(message string) {
	events.Dispatch(NewChatMessage(message))
}

func (m *CountersModule) setupDatabase() error {
	const createCountersTable = `
		CREATE TABLE IF NOT EXISTS counters (
		counter_name TEXT NOT NULL PRIMARY KEY,
		count INTEGER DEFAULT 1,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`

	_, err := m.db.Exec(createCountersTable)
	if err != nil {
		return err
	}

	return nil
}

func (m *CountersModule) incrementCounter(name CounterName) error {
	_, err := m.db.Exec("INSERT INTO counters (counter_name, count) VALUES ($1, 1) ON CONFLICT (counter_name) DO UPDATE SET count = counters.count + 1", name)
	if err != nil {
		logger.Errorf("Failed to increment counter: %s", err.Error())
		return err
	}
	return nil
}

func (m *CountersModule) getCounter(name CounterName) (int, error) {
	const query = `SELECT count FROM counters WHERE counter_name = ?;`
	var counter int
	err := m.db.QueryRow(query, name).Scan(&counter)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

// Register implements events.EventHandler.
func (m *CountersModule) Register() {
	db, err := db.GetDatabase()
	if err != nil {
		logger.Errorf("Failed to open database: %s", err.Error())
		return
	}
	m.db = db
	err = m.setupDatabase()
	if err != nil {
		logger.Errorf("Failed to setup database: %s", err.Error())
		return
	}

	events.Register(m.handleKasamadaEvent)
	events.Register(m.handleKasamadaMessage)
	events.Register(m.handleCommand)
}

func (m *CountersModule) handleCommand(msg *twitch.PrivateMessage) error {
	if !strings.HasPrefix(msg.Message, "!") {
		return nil
	}
	counter, err := intoCounterName(strings.TrimPrefix(strings.ToLower(msg.Message), "!"))
	if err != nil {
		return nil
	}

	count, err := m.getCounter(counter)
	if err != nil {
		return err
	}

	events.Dispatch(NewChatMessage(fmt.Sprintf("%s: %d", counter, count)))

	return nil
}

func (m *CountersModule) handleKasamadaMessage(msg *twitch.PrivateMessage) error {
	if _, ok := msg.User.Badges["broadcaster"]; !ok {
		return nil
	}

	if !strings.HasPrefix(msg.Message, "!add-kasamada") {
		return nil
	}

	err := m.incrementCounter(KasamadasCount)
	if err != nil {
		logger.Errorf("Failed to increment counter: %s", err.Error())
		return err
	}

	count, err := m.getCounter(KasamadasCount)
	if err != nil {
		return err
	}

	m.sendTwitchChatMessage(fmt.Sprintf("%s incrementado. Total de %d", KasamadasCount, count))

	return nil
}

func (m *CountersModule) handleKasamadaEvent(event *MacropadEvent) error {
	if event.Key != "D" {
		return nil
	}

	err := m.incrementCounter(KasamadasCount)
	if err != nil {
		logger.Errorf("Failed to increment counter: %s", err.Error())
		return err
	}

	count, err := m.getCounter(KasamadasCount)
	if err != nil {
		return err
	}

	m.sendTwitchChatMessage(fmt.Sprintf("%s incrementado. Total de %d", KasamadasCount, count))

	return nil
}

var _ events.EventHandler = &CountersModule{}
