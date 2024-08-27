package omegastrikers

import (
	"fmt"
	"strings"

	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// LogLoadingScreen: Loading screen showing: 1. Reason: We have pending travel (the TravelURL is not empty)
//
// LogPMPlayerState: APMPlayerState::OnOwnerOnlyRep_MatchPhaseChangesListString - Applying Index '1' by calling APMGameState::PerformCurrentMatchPhaseEvent(EMatchPhase::PreGame, EMatchPhase::ArenaOverview)
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::PreGame] Current[EMatchPhase::ArenaOverview]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::ArenaOverview] Current[EMatchPhase::CharacterPreSelect]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::CharacterPreSelect] Current[EMatchPhase::BanSelect]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::BanSelect] Current[EMatchPhase::BanCelebration]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::BanCelebration] Current[EMatchPhase::CharacterSelect]
//
// LogPMGameState:   Training Class 'TD_PrimaryAbilityCooldownReduction' MaxPicks '-1'
// LogPMGameState:   Training Class 'TD_StackingSize' MaxPicks '-1'
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::CharacterSelect] Current[EMatchPhase::LoadoutSelect]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::LoadoutSelect] Current[EMatchPhase::VersusScreen]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::VersusScreen] Current[EMatchPhase::FaceOffIntro]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::FaceOffIntro] Current[EMatchPhase::FaceOffCountdown]
//
// LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::FaceOffCountdown] Current[EMatchPhase::InGame]

var (
	LogLineLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"WindowFocus", `LogPMPlayerControllerBase: APMPlayerControllerBase::WindowFocusChanged - Windows Focus Changed!`},
		{"MatchPhaseChangeStart", `LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous\[EMatchPhase::`},
		{"MatchPhaseChangeMid", `\] Current\[EMatchPhase::`},

		{"TrainingClassStart", `LogPMGameState:[ \t]*Training Class '`},
		{"TrainingClassEnd", `' MaxPicks '`},

		{"Comment", `(?:#|//)[^\n]*\n?`},
		{"Ident", `[a-zA-Z_]\w*`},
		{"Int", `\d+`},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
		{"Whitespace", `[ \t\r\n]+`},
	})

	LogLineParser = participle.MustBuild[LogLine](
		participle.Lexer(LogLineLexer),
		participle.Elide("Whitespace", "Comment", "Punct"),
		participle.UseLookahead(2),
	)
)

type MatchState string

var (
	MatchStateOther              MatchState = "Other"
	MatchStatePreGame            MatchState = "PreGame"
	MatchStateArenaOverview      MatchState = "ArenaOverview"
	MatchStateCharacterPreSelect MatchState = "CharacterPreSelect"
	MatchStateBanSelect          MatchState = "BanSelect"
	MatchStateBanCelebration     MatchState = "BanCelebration"
	MatchStateCharacterSelect    MatchState = "CharacterSelect"
	MatchStateLoadoutSelect      MatchState = "LoadoutSelect"
	MatchStateVersusScreen       MatchState = "VersusScreen"
	MatchStateFaceOffIntro       MatchState = "FaceOffIntro"
	MatchStateFaceOffCountdown   MatchState = "FaceOffCountdown"
	MatchStateInGame             MatchState = "InGame"
	MatchStateGoalScore          MatchState = "GoalScore"
	MatchStateGoalCelebration    MatchState = "GoalCelebration"
	MatchStatePostGameSummary    MatchState = "PostGameSummary"
)

type LogLine struct {
	Time              *int                    `'[' @Int '.' Int '.' Int '-' Int '.' Int '.' Int ':' Int ']'`
	Number            *int                    `'[' @Int ']'`
	WindowFocusChange *windowFocusChange      `( @@`
	TrainingClass     *trainingClassAvailable `| @@`
	MatchPhaseChange  *matchPhaseChange       `| @@ )`
}

type windowFocusChange struct {
	Value int `WindowFocus @Int`
}

type matchPhaseChange struct {
	Previous MatchState `MatchPhaseChangeStart @Ident`
	Current  MatchState `MatchPhaseChangeMid @Ident ']'`
}

type trainingClassAvailable struct {
	Class string `TrainingClassStart @Ident TrainingClassEnd Int "'" Whitespace*`
}

type OmegaStrikersEventType string

const (
	OSEventMatchPhaseChange  OmegaStrikersEventType = "match_phase_change"
	OSEventAvailableTraining OmegaStrikersEventType = "available_training"
	OSEventWindowFocus       OmegaStrikersEventType = "window_focus"
)

type OmegaStrikersLogEvent struct {
	RawLine string
	Type    OmegaStrikersEventType
	Log     *LogLine
}

func ParseOmegaStrikersLog(line string) (*OmegaStrikersLogEvent, error) {
	parsed, err := LogLineParser.ParseString("", line)
	if err != nil {
		if strings.Contains(line, "Training Class") {
			logger.Debugf("Failed to parse line: '%s'", line)
		}
		return nil, err
	}

	var kind OmegaStrikersEventType
	if parsed.MatchPhaseChange != nil {
		kind = OSEventMatchPhaseChange
	} else if parsed.WindowFocusChange != nil {
		kind = OSEventWindowFocus
	} else if parsed.TrainingClass != nil {
		kind = OSEventAvailableTraining
	}

	if kind == "" {
		return nil, fmt.Errorf("Unknown event type")
	}

	return &OmegaStrikersLogEvent{
		RawLine: line,
		Type:    kind,
		Log:     parsed,
	}, nil
}
