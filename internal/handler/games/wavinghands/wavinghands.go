package wavinghands

const (
	minTeams = 2
	maxTeams = 6
)

type (
	Hand struct {
		Sequence string `json:"sequence"`
	}
	Spell struct {
		Name        string `json:"name"`
		Sequence    string `json:"sequence"`
		ShSequence  string `json:"sh-sequence"`
		Description string `json:"description"`
		Usage       string `json:"usage"`
		Damage      int    `json:"damage"`
		Resistances string `json:"resistances"`
		Protections string `json:"protections"`
	}
	Living struct {
		HitPoints int    `json:"hp"`
		Wards     string `json:"wards"`
	}
	Monster struct {
		Type        string `json:"type"`
		Living      Living `json:"living"`
		Curses      string `json:"curses"`
		Protections string `json:"protections"`
	}
	Wizard struct {
		Right       Hand    `json:"right"`
		Left        Hand    `json:"left"`
		Name        string  `json:"name"`
		Living      Living  `json:"living"`
		Curses      string  `json:"curses"`
		Protections string  `json:"protections"`
		Monsters    Monster `json:"monsters"`
	}
)

func Remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func GetMaxTeams() int {
	return maxTeams
}
func GetMinTeams() int {
	return minTeams
}
