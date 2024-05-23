package modules

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/db"
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/PuerkitoBio/goquery"
	obsEvents "github.com/andreykaipov/goobs/api/events"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
)

var ThemeNotFound = errors.New("Theme not found")

type UserThemeModule struct {
	db              *db.Database
	client          *twitch.Client
	helix           *helix.Client
	messagedAlready map[string]struct{}
	channel         string
}

func NewUserThemeModule(channel string) *UserThemeModule {
	db, err := db.GetDatabase()
	if err != nil {
		logger.Errorf("Failed to open database: %s", err.Error())
	}

	module := &UserThemeModule{
		db:              db,
		client:          &twitch.Client{},
		messagedAlready: make(map[string]struct{}),
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

	const createUsedThemeAlreadyTable = `
		CREATE TABLE IF NOT EXISTS used_theme_already (
		userID TEXT PRIMARY KEY NOT NULL);`

	_, err := m.db.Exec(createThemesTable)
	if err != nil {
		return err
	}

	_, err = m.db.Exec(createBlacklistTable)
	if err != nil {
		return err
	}

	_, err = m.db.Exec(createUsedThemeAlreadyTable)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserThemeModule) hasUsedThemeAlready(userID string) bool {
	if _, exists := m.messagedAlready[userID]; exists {
		return true
	}
	if m.db == nil {
		return false
	}
	row := m.db.QueryRow("SELECT userID FROM used_theme_already WHERE userID = ?;", userID)
	var uid string
	err := row.Scan(&uid)
	if err != nil {
		return false
	}
	m.messagedAlready[uid] = struct{}{}
	return true
}

func (m *UserThemeModule) setUsedThemeAlready(userID string) {
	if m.db == nil {
		logger.Errorf("Database is not initialized")
		return
	}
	_, err := m.db.Exec("INSERT INTO used_theme_already (userID) VALUES (?) ON CONFLICT(userID) DO NOTHING;", userID)
	if err != nil {
		logger.Errorf("Failed to set used theme already: %s", err.Error())
	}
	m.messagedAlready[userID] = struct{}{}
}

func (m *UserThemeModule) deleteUsedThemeAlready(userID string) {
	delete(m.messagedAlready, userID)
	if m.db == nil {
		logger.Errorf("Database is not initialized")
		return
	}
	_, err := m.db.Exec("DELETE FROM used_theme_already WHERE userID = ?;", userID)
	if err != nil {
		logger.Errorf("Failed to clear used_themes_already table", err.Error())
	}
}

func (m *UserThemeModule) clearUsedThemes() {
	m.messagedAlready = make(map[string]struct{})
	if m.db == nil {
		logger.Errorf("Database is not initialized")
		return
	}
	_, err := m.db.Exec("DELETE FROM used_theme_already")
	if err != nil {
		logger.Errorf("Failed to clear used_themes_already table", err.Error())
	}
}

func (m *UserThemeModule) setTheme(userID, theme string) error {
	if m.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	const insert = `
		INSERT INTO user_themes (userID, theme, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(userID) DO UPDATE SET
			theme=excluded.theme,
			updated_at=excluded.updated_at;`
	_, err := m.db.Exec(insert, userID, theme, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (m *UserThemeModule) setThemeBan(username string) error {
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

func (m *UserThemeModule) isBlacklisted(username string) bool {
	if m.db == nil {
		return false
	}
	const query = `SELECT username FROM user_blacklist WHERE username = ?;`
	row := m.db.QueryRow(query, username)
	var name string
	err := row.Scan(&name)
	return err == nil
}

func (m *UserThemeModule) getTheme(userID string) (string, error) {
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
	events.Register(m.handleTwitchClient)
	events.Register(m.handleHelixTwitchClient)

	events.Register(m.handleTheme)
	events.Register(m.handleChangeTheme)
	events.Register(m.handleReset)
	events.Register(m.handleBan)
	events.Register(m.handleStartStream)
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

func (m *UserThemeModule) handleHelixTwitchClient(client *helix.Client) error {
	m.helix = client
	return nil
}

func (m *UserThemeModule) handleStartStream(e *obsEvents.StreamStateChanged) error {
	if e.OutputActive {
		return nil
	}
	m.clearUsedThemes()
	return nil
}

func (m *UserThemeModule) handleTheme(message *twitch.PrivateMessage) error {
	if !IsSubscriber(&message.User) {
		// ignore non-subscribers
		return nil
	}

	if m.hasUsedThemeAlready(message.User.ID) {
		// User has sent messages already, ignore them
		return nil
	}

	if m.isBlacklisted(message.User.Name) {
		return nil
	}

	theme, err := m.getTheme(message.User.ID)
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

	m.setUsedThemeAlready(message.User.ID)
	events.Dispatch(NewPlayAudioEvent(&resp.Body))
	users, _ := m.helix.GetUsers(&helix.UsersParams{
		IDs: []string{message.User.ID},
	})

	user := users.Data.Users[0]

	events.Dispatch(NewWebEvent("user_theme_played", views.RenderToString(views.MsnNotification(&user, message.User.Color))))

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
	printThemeUsage := func() {
		if m.client != nil {
			m.client.Say(m.channel, "Uso para subs/vips: !tema <url>. A URL deve ser um link de um .mp3 ou um link do www.myinstants.com (max 10s)")
		}
	}

	if message.Message == "!tema" {
		printThemeUsage()
		return nil
	}

	if !strings.HasPrefix(message.Message, "!tema ") {
		return nil
	}

	if !IsSubscriber(&message.User) {
		printThemeUsage()
		return nil
	}

	parts := strings.Split(message.Message, " ")
	if len(parts) < 2 {
		printThemeUsage()
		return nil
	}

	finalURL, err := getThemeFromUrl(parts[1])
	if err != nil || finalURL == "" {
		printThemeUsage()
		return err
	}

	err = m.setTheme(message.User.ID, finalURL)
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
	if len(parts) < 2 {
		return nil
	}
	name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(parts[1]), "@"))

	if name == "all" {
		m.clearUsedThemes()
	}

	m.deleteUsedThemeAlready(message.User.ID)
	logger.Debugf("Reset theme for %s. Users are: %v", name, m.messagedAlready)

	return nil
}

func (m *UserThemeModule) handleBan(message *twitch.PrivateMessage) error {
	if _, exists := message.User.Badges["broadcaster"]; !exists || !strings.HasPrefix(message.Message, "!tema-ban") {
		return nil
	}

	parts := strings.Split(message.Message, " ")
	name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(parts[1]), "@"))

	err := m.setThemeBan(name)
	if err != nil {
		return err
	}

	return nil
}

var _ events.EventHandler = &UserThemeModule{}
