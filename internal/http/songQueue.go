package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Kasama/kasama-twitch-integrations/internal/db"
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/modules"
	mediaplayer "github.com/Kasama/kasama-twitch-integrations/internal/modules/mediaPlayer"
	"github.com/a-h/templ"
	"github.com/beeker1121/goque"
	"github.com/labstack/echo/v4"
)

type SongQueueHandler struct {
	queue *db.Queue[modules.SongQueueItem]
}

func (h *SongQueueHandler) RegisterRoutes(e *echo.Group) {
	e.GET("", h.handleQueue)
	e.GET("/", h.handleQueue)
	e.DELETE("/:id", h.handleDeleteItem)
	e.POST("/control/:action", h.handlePlayerAction)
	e.POST("/:id/swap/:direction", h.handleSwapItem)
	e.GET("/queue", h.handleGetQueue)
}

func NewSongQueueHandler(queue *db.Queue[modules.SongQueueItem]) *SongQueueHandler {
	return &SongQueueHandler{
		queue: queue,
	}
}

func (h *SongQueueHandler) handlePlayerAction(c echo.Context) error {
	action := c.Param("action")
	switch action {
	case "play":
		events.Dispatch(mediaplayer.PlayEvent())
	case "next":
		events.Dispatch(mediaplayer.NextEvent())
	case "enqueue":
		url := c.FormValue("url")
		events.Dispatch(mediaplayer.EnqueueEvent(url, "Kasama"))
	}

	return nil
}

func (h *SongQueueHandler) handleSwapItem(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	direction := c.Param("direction")

	if (idInt == 0 && direction == "up") || (idInt == h.queue.Len()-1 && direction != "up") {
		return Render(c, http.StatusOK, views.SongQueue(h.QueueEntries()))
	}
	swapDirection := 1
	if direction == "up" {
		swapDirection = -1
	}

	items := h.queue.RawItems()

	swapItem := items[idInt+swapDirection]
	item := items[idInt]

	newItems := make([]*goque.PriorityItem, len(items))
	for i, it := range items {
		if i == idInt {
			newItems[i] = swapItem
		} else if i == idInt+swapDirection {
			newItems[i] = item
		} else {
			newItems[i] = it
		}
	}

	err = h.queue.Clear()
	if err != nil {
		return err
	}
	for _, item := range newItems {
		h.queue.PushRaw(item.Priority, item.Value)
	}

	return Render(c, http.StatusOK, views.SongQueue(h.QueueEntries()))
}

func (h *SongQueueHandler) handleDeleteItem(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	items := h.queue.RawItems()

	newItems := make([]*goque.PriorityItem, 0, len(items)-1)

	newItems = append(newItems, items[:idInt]...)
	newItems = append(newItems, items[idInt+1:]...)

	err = h.queue.Clear()
	if err != nil {
		return err
	}
	for _, item := range newItems {
		h.queue.PushRaw(item.Priority, item.Value)
	}

	return Render(c, http.StatusOK, views.SongQueue(h.QueueEntries()))
}

func (h *SongQueueHandler) QueueEntries() []templ.Component {
	items := h.queue.Items()
	is := make([]templ.Component, 0, len(items))
	for i, item := range items {
		is = append(is, views.SongQueueEntry(fmt.Sprintf("%d", i), item.Track.Artists[0].Name, item.Track.Name, item.OriginalQuery))
	}
	return is
}

func (h *SongQueueHandler) handleGetQueue(c echo.Context) error {
	c.Set("skip-log", true)
	return Render(c, http.StatusOK, views.SongQueue(h.QueueEntries()))
}


func (h *SongQueueHandler) handleQueue(c echo.Context) error {

	if h.queue == nil {
		return Render(c, http.StatusTooEarly, views.Page("Twitch Song Queue", views.NotYetSongQueue()))
	}

	is := h.QueueEntries()

	return Render(c, http.StatusOK, views.Page("Twitch Song Queue", views.SongQueuePage(is)))
}
