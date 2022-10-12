package spells

import (
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"golang.org/x/exp/slices"
	"strings"
)

type CureHeavyWounds struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (cHW CureHeavyWounds) cast(wizard *wavinghands.Wizard, target *wavinghands.Living) error {
	if strings.HasSuffix(wizard.Right.Sequence, cHW.Sequence) || strings.HasSuffix(wizard.Left.Sequence, cHW.Sequence) {
		wards := strings.Split(target.Wards, ",")
		wards = append(wards, "cureHeavyWounds") // Lasts one round
		target.Wards = strings.Join(wards, ",")
	}

	return nil
}

func (cHW CureHeavyWounds) clear(target *wavinghands.Living) error {
	wards := strings.Split(target.Wards, ",")
	idx := slices.Index(wards, "cureHeavyWounds")
	wavinghands.Remove(wards, idx)
	target.Wards = strings.Join(wards, ",")
	return nil
}
