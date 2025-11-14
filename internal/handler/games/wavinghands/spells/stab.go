package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
)

type Stab struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (s Stab) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	// Stab is represented as "1" in the sequence and can only be done with one hand
	if (len(wizard.Right.Sequence) > 0 && wizard.Right.Sequence[len(wizard.Right.Sequence)-1] == '1') ||
		(len(wizard.Left.Sequence) > 0 && wizard.Left.Sequence[len(wizard.Left.Sequence)-1] == '1') {

		// Stab is blocked by shield
		if wavinghands.HasShield(target) {
			return fmt.Sprintf("%s tried to stab %s but was blocked by a shield", wizard.Name, target.Selector), nil
		} else {
			target.HitPoints -= 1
			return fmt.Sprintf("%s stabbed %s with a knife", wizard.Name, target.Selector), nil
		}
	}

	return "", nil
}

func GetStabSpell(s *wavinghands.Spell, e error) (*Stab, error) {
	if e != nil {
		return &Stab{}, e
	}

	return &Stab{
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
