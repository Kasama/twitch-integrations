package http

import (
	"net/http"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/modules"
	"github.com/labstack/echo/v4"
)

func HandleMacropad(c echo.Context) error {
	key := c.Param("id")

	err := c.NoContent(http.StatusOK)

	if key != "" {
		events.Dispatch(&modules.MacropadEvent{
			Key:   key,
			State: modules.MacropadEventStatePressed,
		})
	}

	return err
}
