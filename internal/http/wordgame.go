package http

import (
	"net/http"

	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/labstack/echo/v4"
)

type WordGameHandler struct {
}

func NewWordGameHandler() *WordGameHandler {
	return &WordGameHandler{}
}

func (h *WordGameHandler) handleIndex(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("WordGame", views.WordGameStuff()))
}
