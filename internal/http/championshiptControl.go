package http

import (
	"net/http"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/championship"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/Kasama/kasama-twitch-integrations/internal/modules"
	"github.com/labstack/echo/v4"
)

type ChampionshipControlHandler struct {
	championship *championship.Championship
}

func NewChampionshipControlHandler() *ChampionshipControlHandler {
	return &ChampionshipControlHandler{
		championship: championship.NewChampionship(),
	}
}

func (h *ChampionshipControlHandler) RegisterRoutes(e *echo.Group) {
	e.GET("", h.handleControlPanel)
	e.GET("/", h.handleControlPanel)
	e.GET("/overlay", h.handleOverlay)
	e.POST("/setupMatch", h.handleSetupTeam)
	e.POST("/clearMatch", h.handleClearMatch)
	e.GET("/starting", h.handleStarting)
}

func (h *ChampionshipControlHandler) handleControlPanel(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("Championship Control", views.ChampionshipControlPage(h.championship)))
}

func (h *ChampionshipControlHandler) handleOverlay(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("championship overlay", views.ChampionshipOverlay()))
}

func (h *ChampionshipControlHandler) handleClearMatch(c echo.Context) error {
	h.championship.CurrentMatch = nil

	events.Dispatch(modules.NewWebEvent("current_match", views.RenderToString(views.CurrentMatchOverlay(h.championship.CurrentMatch))))
	return Render(c, http.StatusOK, views.Page("Championship Control", views.ChampionshipControlPage(h.championship)))
}

func (h *ChampionshipControlHandler) handleStarting(c echo.Context) error {
	c.Set("skip-log", true)
	t := time.Until(time.Unix(1757790000, 0))
	var s string
	if t < 0 {
		s = "breve"
	} else {
		s = t.Round(time.Second).String()
	}
	return Render(c, http.StatusOK, views.Page("Starting", views.ChampionshipWaiting(s)))
}

func (h *ChampionshipControlHandler) handleSetupTeam(c echo.Context) error {
	teamAName := c.FormValue("teamA")
	teamBName := c.FormValue("teamB")

	logger.Debugf("Setting up match between %s and %s", teamAName, teamBName)

	if teamAName == "" || teamBName == "" {
		return Render(c, http.StatusBadRequest, views.Page("championship control", views.ChampionshipControlPage(h.championship)))
	}

	h.championship.CurrentMatch = &championship.Match{
		TeamA:  h.championship.AvailableTeams[teamAName],
		TeamB:  h.championship.AvailableTeams[teamBName],
		ScoreA: 0,
		ScoreB: 0,
	}

	events.Dispatch(modules.NewWebEvent("current_match", views.RenderToString(views.CurrentMatchOverlay(h.championship.CurrentMatch))))

	return Render(c, http.StatusOK, views.Page("Championship Control", views.ChampionshipControlPage(h.championship)))
}
