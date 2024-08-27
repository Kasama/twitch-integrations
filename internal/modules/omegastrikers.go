package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/Kasama/kasama-twitch-integrations/internal/omegastrikers"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/general"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	// "github.com/andreykaipov/goobs/api/requests/scenes"
)

const AwakeningPath = "/Files/Images/Omega Strikers/Omega Strikers Media Assets/Gear and Awakenings/"

var AwakeningsMap = map[string]string{
	"TD_TakeDownReduceCooldowns":         "AdrenalineRush.png",
	"TD_FasterProjectiles2":              "Aerials.png",
	"TD_SizeIncrease2":                   "BigFish.png",
	"TD_StaggerPowerConversion":          "BulkUp.png",
	"TD_BuffAndDebuffDuration":           "CastToLast.png",
	"TD_EnergyCatalyst":                  "Catalyst.png",
	"TD_FasterDashes2":                   "Chronoboost.png",
	"TD_DistancePower":                   "DeadEye.png",
	"TD_FasterDashes3":                   "Explosive_Entrance.png",
	"TD_SpecialCooldownAfterRounds":      "ExtraSpecial.png",
	"TD_AvoidDamageHitHarder":            "GlassCannon.png",
	"TD_MultiHitsReduceCooldowns":        "HeavyImpact.png",
	"TD_HitRockCooldown":                 "HotShot.png",
	"TD_SizePowerConversion":             "Might_Of_The_Colossus.png",
	"TD_FasterProjectiles":               "Missile Propulsion.png",
	"TD_CreationSize":                    "Monumentalist.png",
	"TD_ComboATarget":                    "OneTwoPunch.png",
	"TD_EnhancedOrbsSpeed":               "OrbDancer.png",
	"TD_EnhancedOrbsCooldown":            "OrbPonderer.png",
	"TD_StaggerSpeedConversion":          "Peak Performance.png",
	"TD_HitsReduceCooldowns":             "PerfectForm.png",
	"TD_PrimaryEcho":                     "PrimeTime.png",
	"TD_KOKing":                          "PrizeFighter.png",
	"TD_StrikeCooldownReduction":         "QuickStrike.png",
	"TD_StackingSize":                    "Rampage.png",
	"TD_PrimaryAbilityCooldownReduction": "RapidFire.png",
	"TD_BaseStaggerAndRegen":             "Reptile Regeneration.png",
	"TD_StaggerCooldownRateConversion":   "Reverberation.png",
	"TD_FasterProjectiles3":              "Siege_Machine.png",
	"TD_HitsIncreaseSpeedAndPower":       "StacksOnStacks.png",
	"TD_HitEnemyBurnThem":                "Stinger.png",
	"TD_StrikeRockTowardsAllies":         "Team_Player.png",
	"TD_MovementAbilityCharges":          "TwinDrive.png",
	"TD_ResistFirstHit":                  "Unstoppable.png",
}

type OmegaStrikersModule struct {
	enabledTrainings []string
	currentGameState omegastrikers.MatchState
	obsClient        *goobs.Client
}

func NewOmegaStrikersModule() *OmegaStrikersModule {
	return &OmegaStrikersModule{
		enabledTrainings: make([]string, 0, 2),
		currentGameState: omegastrikers.MatchStateOther,
	}
}

// Register implements events.EventHandler.
func (m *OmegaStrikersModule) Register() {
	events.Register(m.handleOSEvent)
	events.Register(m.handleOBSClient)
}

func (m *OmegaStrikersModule) handleOBSClient(client *goobs.Client) error {
	m.obsClient = client
	return nil
}

func (m *OmegaStrikersModule) handleOSEvent(ev *omegastrikers.OmegaStrikersLogEvent) error {
	if ev.Type == omegastrikers.OSEventWindowFocus {
		return nil
	}
	if ev.Type == omegastrikers.OSEventMatchPhaseChange {

		///////////////////////////////////////////////////////
		return nil
		///////////////////////////////////////////////////////

		if ev.Log.MatchPhaseChange.Previous == omegastrikers.MatchStateArenaOverview {
			scene := "Estrelas Nascentes - Selection"
			_, _ = m.obsClient.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
				SceneName: &scene,
			})
		}

		// resp, err := m.obsClient.General.GetHotkeyList(&general.GetHotkeyListParams{})
		// if err != nil {
		// 	logger.Debugf("Error getting hotkeys: %s", err)
		// } else {
		// 	logger.Debugf("Hotskeys hotkey: %+v", resp)
		// }

		if ev.Log.MatchPhaseChange.Previous == omegastrikers.MatchStateGoalScore {
			hotkeyName := "instant_replay.trigger"
			resp, err := m.obsClient.General.TriggerHotkeyByName(&general.TriggerHotkeyByNameParams{
				HotkeyName: &hotkeyName,
			})
			if err != nil {
				logger.Debugf("Error triggering hotkey: %s", err)
			} else {
				logger.Debugf("Triggered hotkey: %+v", resp)
			}
		}

		m.currentGameState = ev.Log.MatchPhaseChange.Current
		if m.currentGameState == omegastrikers.MatchStateVersusScreen {
			m.enabledTrainings = make([]string, 0, 2)
			events.Dispatch(NewWebEvent("current_awakenings", views.RenderToString(views.RenderAwakenings(m.enabledTrainings))))
		}
	} else if ev.Type == omegastrikers.OSEventAvailableTraining {
		logger.Debugf("Got training: %s during phase %s", AwakeningsMap[ev.Log.TrainingClass.Class], m.currentGameState)
		if m.currentGameState == omegastrikers.MatchStateArenaOverview || m.currentGameState == omegastrikers.MatchStatePreGame {
			m.enabledTrainings = append(m.enabledTrainings, AwakeningsMap[ev.Log.TrainingClass.Class])
			if len(m.enabledTrainings) >= 2 {
				events.Dispatch(NewWebEvent("current_awakenings", views.RenderToString(views.RenderAwakenings(m.enabledTrainings))))
			}
		}
	}

	return nil
}

var _ events.EventHandler = &OmegaStrikersModule{}
