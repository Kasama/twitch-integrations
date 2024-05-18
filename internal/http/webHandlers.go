package http

import (
	"net/http"

	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/labstack/echo/v4"
)

func HandleIndex(c echo.Context) error {
	logger.Debugf("Handling index")
	return Render(c, http.StatusOK, views.Page("Kasama Twitch Helper", views.Index()))
}

func HandleObsBackground(c echo.Context) error {
	return Render(c, http.StatusOK, views.WavesBackground())
}

func HandleSSEUI(c echo.Context) error {
	return Render(c, http.StatusOK, views.Page("OBS Overlay", views.OBSOverlay()))
}
