package spells

import (
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type CauseHeavyWounds struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (cHW CauseHeavyWounds) cast(wizard *wavinghands.Wizard, target *wavinghands.Living, opponent *wavinghands.Wizard) error {
	if strings.HasSuffix(wizard.Right.Sequence, cHW.Sequence) || strings.HasSuffix(wizard.Left.Sequence, cHW.Sequence) {
		if strings.Contains(target.Wards, "cureHeavyWounds") {
			target.HitPoints -= 1
		} else {
			target.HitPoints -= 3
		}
	}

	return nil
}
