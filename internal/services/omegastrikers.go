package services

import (
	"context"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/hpcloud/tail"
)

const OSLogPath = "~/.local/share/Steam/steamapps/compatdata/1869590/pfx/drive_c/users/steamuser/AppData/Local/OmegaStrikers/Saved/Logs/OmegaStrikers.log"

type OmegaStrikersService struct {
	ctx context.Context
}

type OmegaStrikersLogEvent struct {
	RawLine string
}

func NewOmegaStrikersService(ctx context.Context) *OmegaStrikersService {
	return &OmegaStrikersService{
		ctx: ctx,
	}
}

func (s *OmegaStrikersService) Register() {
	go func() {
		t, err := tail.TailFile(OSLogPath, tail.Config{
			ReOpen:      true,
			MustExist:   false,
			Follow:      true,
			MaxLineSize: 0,
			Logger:      tail.DefaultLogger,
		})
		if err != nil {
			logger.Errorf("Failed to tail OmegaStrikers log: %s", err)
			return
		}

		for line := range t.Lines {
				if line == nil || line.Err != nil {
					err := ""
					if line == nil {
						err = "no line given"
					} else {
						err = line.Err.Error()
					}
					logger.Errorf("Error reading OmegaStrikers log: %s", err)
					continue
				}
				parseOmegaStrikersLog(line.Text)
		}

		return
		for {
			select {
			case <-s.ctx.Done():
				return
			case line := <-t.Lines:
				if line == nil || line.Err != nil {
					err := ""
					if line == nil {
						err = "no line given"
					} else {
						err = line.Err.Error()
					}
					logger.Errorf("Error reading OmegaStrikers log: %s", err)
					continue
				}
				parseOmegaStrikersLog(line.Text)
			}
		}
	}()
}

func parseOmegaStrikersLog(line string) *OmegaStrikersLogEvent {
	return &OmegaStrikersLogEvent{
		RawLine: line,
	}
}

var _ events.EventHandler = &OmegaStrikersService{}
