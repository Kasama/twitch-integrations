package http

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func handleWSHotReload(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			err := websocket.Message.Send(ws, "ping")
			if err != nil {
				break
			}

			_ = websocket.Message.Receive(ws, nil)
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
