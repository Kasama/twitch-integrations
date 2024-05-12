package http

import (
	"net/http"

	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/labstack/echo/v4"
)

func HandleIndex(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("Kasama Twitch Helper", views.Index()))
}
