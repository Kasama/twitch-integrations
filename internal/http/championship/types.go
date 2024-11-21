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
			"Grupo de Tudo": {
				Name: "Grupo de Tudo",
				Logo: "",
			},
			"Os come terra": {
				Name: "Os come terra",
				Logo: "",
			},
			"dispara ervilha": {
				Name: "dispara ervilha",
				Logo: "",
			},
			"Amamos Hentai": {
				Name: "Amamos Hentai",
				Logo: "",
			},
			"Eu amo Dandadan": {
				Name: "Eu amo Dandadan",
				Logo: "",
			},
			"Imperadores das Chamas": {
				Name: "Imperadores das Chamas",
				Logo: "",
			},
			"It's the end": {
				Name: "It's the end",
				Logo: "",
			},
			"Mains juno sem juno": {
				Name: "Mains juno sem juno",
				Logo: "",
			},
			"353 pipocas": {
				Name: "353 pipocas",
				Logo: "",
			},
		},
		CurrentMatch: nil,
	}
}
