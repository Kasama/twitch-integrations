package http

import (
	"fmt"
	"net/http"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/Kasama/kasama-twitch-integrations/internal/modules"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Handlers struct {
	server       *echo.Echo
	logger       logger.Logger
	environment  string
	twitchConfig *twitch.TwitchConfig
	twitchAuth   *twitch.TwitchAuth
	webEvents    *modules.WebEventsModule
}

type State struct {
}

func Render(c echo.Context, statusCode int, t templ.Component) error {
	c.Response().Writer.WriteHeader(statusCode)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(c.Request().Context(), c.Response().Writer)
}

func NewHandlers(env string, twitchConfig *twitch.TwitchConfig, webEvents *modules.WebEventsModule) *Handlers {
	return &Handlers{
		server:       echo.New(),
		logger:       logger.New("twitch_helper", log.DEBUG),
		environment:  env,
		twitchConfig: twitchConfig,
		webEvents:    webEvents,
	}
}

func (h *Handlers) loadCookieAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth, err := TwitchAuthFromCookies(c, h.twitchConfig)
		if err == nil {
			if h.twitchAuth == nil {
				events.Dispatch(auth)
			}
			h.twitchAuth = auth
		}
		return next(c)
	}
}

func (h *Handlers) updateCookieAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if h.twitchAuth != nil {
			SaveToCookies(c, h.twitchAuth)
		}
		return err
	}
}

func (h *Handlers) Start(address string, port string) error {
	h.RegisterRoutes()
	return h.server.Start(fmt.Sprintf("%s:%s", address, port))
}

func (h *Handlers) RegisterRoutes() {
	h.server.Use(getLoggerMiddleware(h.logger))
	h.server.Use(h.loadCookieAuth)
	h.server.Use(h.updateCookieAuth)
	h.server.Use(middleware.Recover())

	twitchHandler := NewTwitchHandler(h.twitchConfig)

	// API routes
	h.server.GET("/api/livez", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })
	h.server.GET("/api/sse", h.webEvents.HandleSSE)
	h.server.GET("/ws/dev/hot-reload", handleWSHotReload)

	// Web UI routes
	h.server.Static("/assets", "assets")
	h.server.GET("/", HandleIndex)
	h.server.GET("/auth/twitch", twitchHandler.handleAuth)
	h.server.GET("/auth/twitch/redirect", twitchHandler.handleRedirect)
	h.server.GET("/twitch", twitchHandler.handleIndex)
	h.server.GET("/obs/background", HandleObsBackground)
	h.server.GET("/sse", HandleSSEUI)
	h.server.GET("/obsOverlay", HandleSSEUI)

	h.server.Logger.Debug("Routes registered")
	h.server.Logger.Info(h.server.Routes())
}

func getLoggerMiddleware(logger logger.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError: true,
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if c.Path() == "/api/livez" {
				return nil
			}
			logger.
				WithField("status", v.Status).
				WithField("error", v.Error).
				WithField("method", v.Method).
				WithField("uri", v.URI).
				Info("request")
			return nil
		},
	})
}
