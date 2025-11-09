package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type CauseLightWounds struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (cLW CauseLightWounds) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if blocked, msg := wavinghands.CounterSpellBlocks(target, wizard.Name, cLW.Name); blocked {
		return msg, nil
	}

	if (len(wizard.Right.Sequence) >= len(cLW.Sequence) && strings.HasSuffix(wizard.Right.Sequence, cLW.Sequence)) ||
		(len(wizard.Left.Sequence) >= len(cLW.Sequence) && strings.HasSuffix(wizard.Left.Sequence, cLW.ShSequence)) {

		if wavinghands.HasWard(target, "cureLightWounds") {
			wavinghands.RemoveWard(target, "cureLightWounds")
			target.HitPoints -= 1
			return fmt.Sprintf("%s caused light wounds on %s but they were protected and only sustained minimal damage", wizard.Name, target.Selector), nil
		} else {
			target.HitPoints -= 2
			return fmt.Sprintf("%s caused light wounds on %s", wizard.Name, target.Selector), nil
		}
	}

	return "", nil
}

func GetCauseLightWoundsSpell(s *wavinghands.Spell, e error) (*CauseLightWounds, error) {
	if e != nil {
		return &CauseLightWounds{}, e
	}

	return &CauseLightWounds{
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
