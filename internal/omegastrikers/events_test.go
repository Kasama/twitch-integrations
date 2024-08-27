package omegastrikers_test

import (
	"testing"

	"github.com/Kasama/kasama-twitch-integrations/internal/omegastrikers"
)

func TestParseOmegaStrikersLog(t *testing.T) {
	// testLine := "[2024.08.10-20.27.42:084][862]LogPMPlayerControllerBase: APMPlayerControllerBase::WindowFocusChanged - Windows Focus Changed! 0"
	testLine := "[2024.08.07-01.46.28:362][738]LogPMGameState: Display: APMGameState::PerformCurrentMatchPhaseEvents - Previous[EMatchPhase::CharacterSelect] Current[EMatchPhase::LoadoutSelect]"
	// testLine := "[2024.08.07-01.46.28:062][738]LogPMGameState:   Training Class 'TD_PrimaryAbilityCooldownReduction' MaxPicks '-1'"
	// testLine := "[2024.08.07-05.35.34:179][208]LogPMGameState:   Training Class 'TD_FasterDashes3' MaxPicks '1'"
	// testLine := "[2024.08.08-04.35.21:257][396]LogPMGameState:   Training Class 'TD_HitsIncreaseSpeedAndPower' MaxPicks '-1'\r\n"
	// testLine := "[2024.08.10-20.09.51:532][629]LogPMGameState:   Training Class 'TD_MultiHitsReduceCooldowns' MaxPicks '1'"
	// testLine := "[2024.08.10-20.47.57:683][236]LogPMGameState: \tTraining Class 'TD_StrikeRockTowardsAllies' MaxPicks '1'\r"

	// investigate "LogPMPerfStatsSubsystem: Game context"

	l, err := omegastrikers.ParseOmegaStrikersLog(testLine)
	if err != nil {
		t.Errorf("Failed to parse line: %+v", err)
	}
	t.Logf("Parsed line: %+v", l)

	// if l.Time == nil || *(l.Time) != 1 {
	// t.Errorf("Expected Time to be 1, got %v", l.Time)
	// }

}
