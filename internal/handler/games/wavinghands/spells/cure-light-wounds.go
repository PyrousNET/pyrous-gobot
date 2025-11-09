package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type CureLightWounds struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (cLW CureLightWounds) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if (len(wizard.Right.Sequence) >= len(cLW.Sequence) && strings.HasSuffix(wizard.Right.Sequence, cLW.Sequence)) ||
		(len(wizard.Left.Sequence) >= len(cLW.Sequence) && strings.HasSuffix(wizard.Left.Sequence, cLW.ShSequence)) {

		wavinghands.AddWard(target, "cureLightWounds")

		// Heal 2 points but don't exceed starting health (typically 15)
		target.HitPoints += 2
		if target.HitPoints > 15 {
			target.HitPoints = 15
		}

		return fmt.Sprintf("%s has cast Cure Light Wounds on %s", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetCureLightWoundsSpell(s *wavinghands.Spell, e error) (*CureLightWounds, error) {
	if e != nil {
		return &CureLightWounds{}, e
	}

	return &CureLightWounds{
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

func (cLW CureLightWounds) clear(target *wavinghands.Living) error {
	wavinghands.RemoveWard(target, "cureLightWounds")
	return nil
}
