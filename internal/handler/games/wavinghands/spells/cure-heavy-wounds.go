package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
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

func (cHW CureHeavyWounds) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	rightHandMatch := len(wizard.Right.Sequence) >= len(cHW.Sequence) && strings.HasSuffix(wizard.Right.Sequence, cHW.Sequence)
	leftHandMatch := cHW.ShSequence != "" && len(wizard.Left.Sequence) >= len(cHW.ShSequence) && strings.HasSuffix(wizard.Left.Sequence, cHW.ShSequence)

	if rightHandMatch || leftHandMatch {
		wavinghands.AddWard(target, "cureHeavyWounds") // Lasts one round

		return fmt.Sprintf("%s has cast Cure Heavy Wounds on %s", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetCureHeavyWoundsSpell(s *wavinghands.Spell, e error) (*CureHeavyWounds, error) {
	if e != nil {
		return &CureHeavyWounds{}, e
	}

	return &CureHeavyWounds{
		Name:        s.Name,
		Sequence:    s.Sequence,
		ShSequence:  s.ShSequence,
		Description: s.Description,
		Usage:       s.Usage,
		Damage:      s.Damage,
		Resistences: s.Resistances,
		Protections: s.Protections,
	}, nil
}

func (cHW CureHeavyWounds) clear(target *wavinghands.Living) error {
	wavinghands.RemoveWard(target, "cureHeavyWounds")
	return nil
}
