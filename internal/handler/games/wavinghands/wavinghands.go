package wavinghands

const (
	minTeams = 2
	maxTeams = 6
)

type (
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
	Monster struct {
		Type        string `json:"type"`
		HitPoints   int    `json:"hp"`
		Curses      string `json:"curses"`
		Protections string `json:"protections"`
	}

	Wizard struct {
		WhPlaying   string  `json:"wh-playing"`
		Name        string  `json:"name"`
		HitPoints   int     `json:"hp"`
		Curses      string  `json:"curses"`
		Protections string  `json:"protections"`
		Monsters    Monster `json:"monsters"`
	}

	GameData struct {
	}
)
