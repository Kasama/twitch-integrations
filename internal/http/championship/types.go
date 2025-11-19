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
			"Dragões do Midfield": {
				Name: "Dragões do Midfield",
				Logo: "/campAssets/DragoesDoMidfield.png",
			},
			"Cinzenta Fan Clube": {
				Name: "Cinzenta Fan Clube",
				Logo: "",
			},
			"DST | DISTINTOS SEM TENTAÇÃO": {
				Name: "DST",
				Logo: "",
			},
			"Ku Com K": {
				Name: "KCK",
				Logo: "/campAssets/kck.webp",
			},
			"HIV": {
				Name: "HIV",
				Logo: "",
			},
			"Marmotes": {
				Name: "Marmotes",
				Logo: "/campAssets/marmotes.webp",
			},
			"Los Pitufos": {
				Name: "Los Pitufos",
				Logo: "",
			},
			"Amigos do Fut": {
				Name: "Amigos do Fut",
				Logo: "/campAssets/amigosdofut.webp",
			},
			"Play Better Win More | PBWM": {
				Name: "PBWM",
				Logo: "/campAssets/pbwm.webp",
			},
			"Maxi Cocido": {
				Name: "Maxi Cocido",
				Logo: "/campAssets/maxicocido.webp",
			},
		},
		CurrentMatch: nil,
	}
}
