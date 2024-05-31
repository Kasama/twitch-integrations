package http

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

var first bool = true

func handleWSHotReload(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		if first {
			_ = websocket.Message.Send(ws, "reload")
			first = false
		}
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
