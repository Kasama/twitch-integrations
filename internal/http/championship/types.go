package championship

type Championship struct {
	AvailableTeams map[string]*Team
	CurrentMatch   *Match
}

type Team struct {
	Name string
	Logo string
}

type Match struct {
	TeamA  *Team
	TeamB  *Team
	ScoreA int
	ScoreB int
}

func NewChampionship() *Championship {
	return &Championship{
		AvailableTeams: map[string]*Team{
			"Nemesis": {
				Name: "Nemesis",
				Logo: "",
			},
			"Never Punished Squad": {
				Name: "Never Punished Squad",
				Logo: "",
			},
			"Flores Repetentes": {
				Name: "Flores Repetentes",
				Logo: "",
			},
			"Bazingers": {
				Name: "Bazingers",
				Logo: "",
			},
			"Clickbait": {
				Name: "Clickbait",
				Logo: "",
			},
			"Alérgicos a PL": {
				Name: "Alérgicos a PL",
				Logo: "",
			},
			"DDR 2nd Mix": {
				Name: "DDR 2nd Mix",
				Logo: "",
			},
			"Team 8": {
				Name: "Team 8",
				Logo: "",
			},
			"Genios da Bola": {
				Name: "Genios da Bola",
				Logo: "",
			},
			"Dezativado": {
				Name: "Dezativado",
				Logo: "",
			},
			"Bau Bau": {
				Name: "Bau Bau",
				Logo: "",
			},
		},
		CurrentMatch: nil,
	}
}
