package services

import (
	"context"
	"io"

	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/Kasama/kasama-twitch-integrations/internal/omegastrikers"
	"github.com/hpcloud/tail"
)

const OSLogPath = "/home/roberto/.local/share/Steam/steamapps/compatdata/1869590/pfx/drive_c/users/steamuser/AppData/Local/OmegaStrikers/Saved/Logs/OmegaStrikers.log"

type OmegaStrikersService struct {
	ctx context.Context
}

func NewOmegaStrikersService(ctx context.Context) *OmegaStrikersService {
	return &OmegaStrikersService{
		ctx: ctx,
	}
}

func (s *OmegaStrikersService) Register() {
	logger.Debugf("Registering OmegaStrikersService")
	go func() {
		t, err := tail.TailFile(OSLogPath, tail.Config{
			ReOpen:      true,
			MustExist:   false,
			Follow:      true,
			MaxLineSize: 0,
			Logger:      tail.DefaultLogger,
			Location:    &tail.SeekInfo{
				Offset: 0,
				Whence: io.SeekEnd,
			},
		})
		if err != nil {
			logger.Errorf("Failed to tail OmegaStrikers log: %s", err)
			return
		}

		for {
			select {
			case <-s.ctx.Done():
				return
			case line := <-t.Lines:
				if line == nil || line.Err != nil {
					continue
				}
				event, err := omegastrikers.ParseOmegaStrikersLog(line.Text)
				if err != nil {
					continue
				}
				events.Dispatch(event)
			}
		}
	}()
}

var _ events.EventHandler = &OmegaStrikersService{}
