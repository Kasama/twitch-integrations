package http

import (
	"net/http"

	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/labstack/echo/v4"
)

func HandleObsBackground(c echo.Context) error {
	return Render(c, http.StatusOK, views.WavesBackground())
}
