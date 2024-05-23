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
		logger.Errorf("Error connecting to OBS: %s", err)
		return &OBSService{}
	}
	return &OBSService{
		address:   address,
		password:  password,
		obsClient: client,
	}
}

// Register implements events.EventHandler.
func (obs *OBSService) Register() {
	if obs.obsClient == nil {
		return
	}

	events.Dispatch(obs.obsClient)

	exit := make(chan struct{})

	go func() {
		<-exit
		_ = obs.obsClient.Disconnect()
	}()

	go func(c *goobs.Client) {
		c.Listen(func(event any) {
			switch e := event.(type) {
			case *obsEvents.CurrentPreviewSceneChanged:
				events.Dispatch(e)
			case *obsEvents.SceneItemSelected:
				events.Dispatch(e)
			case *obsEvents.StreamStateChanged:
				events.Dispatch(e)
			}
		})
	}(obs.obsClient)
}

var _ events.EventHandler = &OBSService{}
