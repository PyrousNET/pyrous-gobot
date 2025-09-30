package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type Missile struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (m Missile) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if (len(wizard.Right.Sequence) >= len(m.Sequence) && strings.HasSuffix(wizard.Right.Sequence, m.Sequence)) ||
		(len(wizard.Left.Sequence) >= len(m.Sequence) && strings.HasSuffix(wizard.Left.Sequence, m.ShSequence)) {

		// Missile is blocked by shield
		if strings.Contains(target.Wards, "shield") {
			return fmt.Sprintf("%s cast missile at %s but it was blocked by a shield", wizard.Name, target.Selector), nil
		} else {
			target.HitPoints -= 1
			return fmt.Sprintf("%s hit %s with a missile", wizard.Name, target.Selector), nil
		}
	}

	return "", nil
}

func GetMissileSpell(s *wavinghands.Spell, e error) (*Missile, error) {
	if e != nil {
		return &Missile{}, e
	}

	return &Missile{
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