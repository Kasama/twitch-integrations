package http

import (
	"fmt"
	"net/http"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/global"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/Kasama/kasama-twitch-integrations/internal/twitch"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Handlers struct {
	server           *echo.Echo
	logger           logger.Logger
	environment      string
	eventsDispatcher *events.EventDispatcher
	twitchConfig     *twitch.TwitchConfig
}

type State struct {
}

func Render(c echo.Context, statusCode int, t templ.Component) error {
	c.Response().Writer.WriteHeader(statusCode)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(c.Request().Context(), c.Response().Writer)
}

func NewHandlers(env string, twitchConfig *twitch.TwitchConfig, dispatcher *events.EventDispatcher) *Handlers {
	return &Handlers{
		server:           echo.New(),
		logger:           logger.New("twitch_helper", log.DEBUG),
		environment:      env,
		eventsDispatcher: dispatcher,
		twitchConfig:     twitchConfig,
	}
}

func loadCookieAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth, err := TwitchAuthFromCookies(c)
		if err == nil {
			global.Global.SetTwitchAuth(auth)
		}
		return next(c)
	}
}

func updateCookieAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		token := global.Global.GetTwitchToken()
		if token != nil {
			SaveToCookies(c, &twitch.TwitchAuth{Token: token})
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
	h.server.Use(loadCookieAuth)
	h.server.Use(updateCookieAuth)
	h.server.Use(middleware.Recover())

	twitchHandler := NewTwitchHandler(h.twitchConfig, h.eventsDispatcher)

	// Api routes
	h.server.GET("/api/livez", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	h.server.POST("/api/services/twitch/chat", twitchHandler.handleEnableChatService)
	h.server.DELETE("/api/services/twitch/chat", twitchHandler.handleDisableChatService)
	// if h.environment == "development" {
	// 	h.server.GET("/ws/dev/hot-reload", handleWSHotReload)
	// }

	// Web UI routes
	h.server.GET("/", HandleIndex)
	h.server.GET("/auth/twitch", twitchHandler.handleAuth)
	h.server.GET("/auth/twitch/redirect", twitchHandler.handleRedirect)
	h.server.GET("/twitch", twitchHandler.handleIndex)

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
