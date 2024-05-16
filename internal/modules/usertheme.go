package modules

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/PuerkitoBio/goquery"
	"github.com/gempir/go-twitch-irc/v4"
	_ "github.com/mattn/go-sqlite3"
)

const localDatabase = "run/database.db"

var ThemeNotFound = errors.New("Theme not found")

type UserThemeModule struct {
	db              *sql.DB
	messagedAlready map[string]struct{}
	client          *twitch.Client
	channel         string
}

func NewUserThemeModule(channel string) *UserThemeModule {
	db, err := sql.Open("sqlite3", localDatabase)
	if err != nil {
		logger.Errorf("Failed to open database: %s", err.Error())
	}

	module := &UserThemeModule{
		db:              db,
		messagedAlready: map[string]struct{}{},
		client:          &twitch.Client{},
		channel:         channel,
	}
	err = module.setupDatabase()
	if err != nil {
		logger.Errorf("Failed to setup database: %s", err.Error())
	}

	return module
}

func (m *UserThemeModule) setupDatabase() error {
	if m.db == nil {
		return fmt.Errorf("database is not initialized")
	}

	const createThemesTable = `
		CREATE TABLE IF NOT EXISTS user_themes (
		userID TEXT PRIMARY KEY NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		theme TEXT NOT NULL);`

	const createBlacklistTable = `
		CREATE TABLE IF NOT EXISTS user_blacklist (
		username TEXT PRIMARY KEY NOT NULL);`

	_, err := m.db.Exec(createThemesTable)
	if err != nil {
		return err
	}

	_, err = m.db.Exec(createBlacklistTable)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserThemeModule) SetTheme(userID, theme string) error {
	if m.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	const insert = `INSERT INTO user_themes (userID, theme, updated_at) VALUES (?, ?, ?);`
	_, err := m.db.Exec(insert, userID, theme, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (m *UserThemeModule) SetThemeBan(username string) error {
	if m.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	const insert = `INSERT INTO user_blacklist (username) VALUES (?);`
	_, err := m.db.Exec(insert, username)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserThemeModule) IsBlacklisted(username string) bool {
	if m.db == nil {
		return false
	}
	const query = `SELECT username FROM user_blacklist WHERE username = ?;`
	row := m.db.QueryRow(query, username)
	var name string
	err := row.Scan(&name)
	if err != nil {
		return false
	}
	return true
}

func (m *UserThemeModule) GetTheme(userID string) (string, error) {
	if m.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}
	const query = `SELECT theme FROM user_themes WHERE userID = ?;`
	row := m.db.QueryRow(query, userID)
	var theme string
	err := row.Scan(&theme)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ThemeNotFound
		}
		return "", err
	}
	return theme, nil
}

// Register implements events.EventHandler.
func (m *UserThemeModule) Register() {
	events.Register(m.handleTheme)
	events.Register(m.handleChangeTheme)
	events.Register(m.handleReset)
	events.Register(m.handleBan)
	events.Register(m.handleTwitchClient)
}

func IsSubscriber(user *twitch.User) bool {
	_, sub := user.Badges["subscriber"]
	_, founder := user.Badges["founder"]
	return sub || founder
}

func (m *UserThemeModule) handleTwitchClient(client *twitch.Client) error {
	m.client = client
	return nil
}

func (m *UserThemeModule) handleTheme(message *twitch.PrivateMessage) error {
	if !IsSubscriber(&message.User) {
		// ignore non-subscribers
		return nil
	}

	if _, exists := m.messagedAlready[message.User.Name]; exists {
		// User has sent messages already, ignore them
		return nil
	}

	if m.IsBlacklisted(message.User.Name) {
		return nil
	}

	theme, err := m.GetTheme(message.User.ID)
	if err != nil {
		if err == ThemeNotFound {
			return nil
		} else {
			return err
		}
	}

	resp, err := http.Get(theme)
	if err != nil {
		return err
	}

	m.messagedAlready[message.User.Name] = struct{}{}
	events.Dispatch(NewPlayAudioEvent(&resp.Body))

	return nil
}

func getThemeFromUrl(rawurl string) (string, error) {
	finalURL := ""

	rawUrl := strings.TrimSpace(rawurl)
	url, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(url.Path, ".mp3") {
		logger.Debugf("Got mp3 file: %s", rawUrl)
		return rawUrl, nil
	}

	if url.Hostname() == "www.myinstants.com" && strings.Contains(url.Path, "/instant/") {
		logger.Debugf("Got media file: %s", rawUrl)
		resp, err := http.Get(url.String())
		if err != nil {
			return "", err
		}

		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return "", err
		}

		document.Find("#instant-page-button-element").EachWithBreak(func(i int, s *goquery.Selection) bool {
			urlpath, _ := s.Attr("data-url")
			finalURL = "https://www.myinstants.com" + urlpath
			return false
		})
	}

	return finalURL, nil

}

func (m *UserThemeModule) handleChangeTheme(message *twitch.PrivateMessage) error {
	logger.Debugf("message: %+v", message)
	if !IsSubscriber(&message.User) || !strings.HasPrefix(message.Message, "!tema ") {
		return nil
	}

	printThemeUsage := func() {
		logger.Debugf("Got client: %v", m.client)
		if m.client != nil {
			m.client.Say(m.channel, "Uso: !tema <url>. A URL deve ser um link de um .mp3 ou um link do www.myinstants.com (max 10s)")
		}
	}

	parts := strings.Split(message.Message, " ")
	finalURL, err := getThemeFromUrl(parts[1])
	if err != nil || finalURL == "" {
		printThemeUsage()
		return err
	}

	err = m.SetTheme(message.User.ID, finalURL)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserThemeModule) handleReset(message *twitch.PrivateMessage) error {
	if _, exists := message.User.Badges["broadcaster"]; !exists || !strings.HasPrefix(message.Message, "!tema-reset") {
		return nil
	}

	parts := strings.Split(message.Message, " ")
	name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(parts[1]), "@"))

	delete(m.messagedAlready, name)
	logger.Debugf("Reset theme for %s. Users are: %v", name, m.messagedAlready)

	return nil
}

func (m *UserThemeModule) handleBan(message *twitch.PrivateMessage) error {
	if _, exists := message.User.Badges["broadcaster"]; !exists || !strings.HasPrefix(message.Message, "!tema-ban") {
		return nil
	}

	parts := strings.Split(message.Message, " ")
	name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(parts[1]), "@"))

	err := m.SetThemeBan(name)
	if err != nil {
		return err
	}

	return nil
}

var _ events.EventHandler = &UserThemeModule{}
