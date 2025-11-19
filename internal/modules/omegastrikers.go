package modules

import (
	"github.com/Kasama/kasama-twitch-integrations/internal/events"
	"github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	"github.com/Kasama/kasama-twitch-integrations/internal/logger"

	// "github.com/Kasama/kasama-twitch-integrations/internal/http/views"
	// "github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/Kasama/kasama-twitch-integrations/internal/omegastrikers"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/general"
	// "github.com/andreykaipov/goobs/api/requests/scenes"
	// "github.com/andreykaipov/goobs/api/requests/general"
	// "github.com/andreykaipov/goobs/api/requests/scenes"
	// "github.com/andreykaipov/goobs/api/requests/scenes"
)

const AwakeningPath = "/Files/Images/Omega Strikers/Omega Strikers Media Assets/Gear and Awakenings/"

var AwakeningsMap = map[string]string{
	"TD_BlessingShare":                   "100px-Spark_of_Leadership.png",
	"TD_BlessingPower":                   "100px-SparkofStrength.png",
	"TD_BlessingCooldownRate":            "100px-SparkofFocus.png",
	"TD_BlessingSpeed":                   "100px-SparkofAgility.png",
	"TD_BlessingMaxStagger":              "100px-SparkofResilience.png",
	"TD_SizeIncrease":                    "100px-BuiltDifferent.png",
	"TD_SizeIncrease2":                   "100px-BigFish.png",
	"TD_Revive":                          "100px-Recovery_Drone.png",
	"TD_StackingSize":                    "100px-Rampage.png",
	"TD_SizePowerConversion":             "100px-Might_Of_The_Colossus.png",
	"TD_BarrierBuff":                     "100px-Demolitionist.png",
	"TD_StaggerPowerConversion":          "100px-BulkUp.png",
	"TD_StaggerCooldownRateConversion":   "100px-Reverberation.png",
	"TD_StaggerSpeedConversion":          "100px-Peak_Performance.png",
	"TD_CreationSize":                    "100px-Monumentalist.png",
	"TD_CreationSizeLifeTime":            "100px-TimelessCreator.png",
	"TD_FasterProjectiles3":              "100px-Siege_Machine.png",
	"TD_BuffAndDebuffDuration":           "100px-CastToLast.png",
	"TD_FasterDashes2":                   "100px-Chronoboost.png",
	"TD_SpecialCooldownAfterRounds":      "100px-ExtraSpecial.png",
	"TD_EmpoweredHitsBuff":               "100px-Specialized_Training.png",
	"TD_MovementAbilityCharges":          "100px-TwinDrive.png",
	"TD_PrimaryEcho":                     "100px-PrimeTime.png",
	"TD_PrimaryAbilityCooldownReduction": "100px-RapidFire.png",
	"TD_HitsReduceCooldowns":             "100px-PerfectForm.png",
	"TD_HitRockCooldown":                 "100px-HotShot.png",
	"TD_ShrinkSelfGrowAllies":            "100px-Among_Titans.png",
	"TD_StrikeRockTowardsAllies":         "100px-Team_Player.png",
	"TD_OrbShare":                        "100px-OrbReplicator.png",
	"TD_EnhancedOrbsCooldown":            "100px-OrbPonderer.png",
	"TD_EnhancedOrbsSpeed":               "100px-OrbDancer.png",
	"TD_KOKing":                          "100px-PrizeFighter.png",
	"TD_TakeDownReduceCooldowns":         "100px-AdrenalineRush.png",
	"TD_HitEnemyBurnThem":                "100px-Stinger.png",
	"TD_MultiHitsReduceCooldowns":        "100px-HeavyImpact.png",
	"TD_AvoidDamageHitHarder":            "100px-GlassCannon.png",
	"TD_ComboATarget":                    "100px-OneTwoPunch.png",
	"TD_DistancePower":                   "100px-DeadEye.png",
	"TD_FasterProjectiles":               "100px-Missile_Propulsion.png",
	"TD_FasterProjectiles2":              "100px-Aerials.png",
	"TD_FasterDashes3":                   "100px-Explosive_Entrance.png",
	"TD_FasterDashes":                    "100px-SuperSurge.png",
	"TD_EdgePower":                       "100px-KnifesEdge.png",
	"TD_HitSpeed":                        "100px-FightOrFlight.png",
	"TD_StrikeCooldownReduction":         "100px-QuickStrike.png",
	"TD_HitsIncreaseSpeedAndPower":       "100px-StacksOnStacks.png",
	"TD_ResistFirstHit":                  "100px-Unstoppable.png",
	"TD_IncreasedSpeedWithStagger":       "100px-StaggerSwagger.png",
	"TD_BaseStaggerAndRegen":             "100px-Clarion_Corp_Regenerator.png",
	"TD_HitAnythingRestoreStagger":       "100px-TempoSwings.png",
	"TD_EnergyCatalyst":                  "100px-Catalyst.png",
	"TD_EnergyConversion":                "100px-Egoist.png",
	"TD_EnergyDischarge":                 "100px-FireUp.png",
	"TD_KnockAnythingRecoverStagger":     "100px-ViciousVambrace.png",
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
		// return nil
		///////////////////////////////////////////////////////

		// if ev.Log.MatchPhaseChange.Previous == omegastrikers.MatchStateArenaOverview {
		// 	scene := "Estrelas Nascentes - Selection"
		// 	_, _ = m.obsClient.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
		// 		SceneName: &scene,
		// 	})
		// }

		resp, err := m.obsClient.General.GetHotkeyList(&general.GetHotkeyListParams{})
		if err != nil {
			logger.Debugf("Error getting hotkeys: %s", err)
		} else {
			logger.Debugf("Hotskeys hotkey: %+v", resp)
		}

		// if ev.Log.MatchPhaseChange.Previous == omegastrikers.MatchStateGoalScore {
		// 	hotkeyName := "instant_replay.trigger"
		// 	resp, err := m.obsClient.General.TriggerHotkeyByName(&general.TriggerHotkeyByNameParams{
		// 		HotkeyName: &hotkeyName,
		// 	})
		// 	if err != nil {
		// 		logger.Debugf("Error triggering hotkey: %s", err)
		// 	} else {
		// 		logger.Debugf("Triggered hotkey: %+v", resp)
		// 	}
		// }

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
