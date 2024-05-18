package services

import (
	"context"
	"time"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
)

type TimerService struct {
	ctx context.Context
}

func NewTimerService(ctx context.Context) *TimerService {
	return &TimerService{
		ctx: ctx,
	}
}

// Register implements events.EventHandler.
func (s *TimerService) Register() {
	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-s.ctx.Done():
				return
			case t := <-ticker.C:
				events.Dispatch(&t)
			}
		}
	}()
}

var _ events.EventHandler = &TimerService{}
