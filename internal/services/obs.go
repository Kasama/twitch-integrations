package services

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/andreykaipov/goobs"
	obsEvents "github.com/andreykaipov/goobs/api/events"
)

type OBSService struct {
	address   string
	password  string
	obsClient *goobs.Client
}

func NewOBSService(address, password string) *OBSService {
	client, err := goobs.New(address, goobs.WithPassword(password))
	if err != nil {
		return nil
	}
	return &OBSService{
		address:   address,
		password:  password,
		obsClient: client,
	}
}

// Register implements events.EventHandler.
func (obs *OBSService) Register() {
	events.Register(obs.handleCurrentScenePreviewSceneChanged)

	obs.obsClient.Listen(func(event any) {
		switch e := event.(type) {
		case *obsEvents.CurrentPreviewSceneChanged:
			events.Dispatch(e)
		case *obsEvents.SceneItemSelected:
			events.Dispatch(e)
		}
	})
}

func (obs *OBSService) handleCurrentScenePreviewSceneChanged(e *obsEvents.CurrentPreviewSceneChanged) error {
	logger.Debugf("Current preview scene changed: %v", e)
	return nil
}

var _ events.EventHandler = &OBSService{}
