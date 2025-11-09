package spells

import (
	"fmt"
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

func (cHW CauseHeavyWounds) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if blocked, msg := wavinghands.CounterSpellBlocks(target, wizard.Name, cHW.Name); blocked {
		return msg, nil
	}

	rightHandMatch := len(wizard.Right.Sequence) >= len(cHW.Sequence) && strings.HasSuffix(wizard.Right.Sequence, cHW.Sequence)
	leftHandMatch := cHW.ShSequence != "" && len(wizard.Left.Sequence) >= len(cHW.ShSequence) && strings.HasSuffix(wizard.Left.Sequence, cHW.ShSequence)

	if rightHandMatch || leftHandMatch {
		if wavinghands.HasWard(target, "cureHeavyWounds") {
			wavinghands.RemoveWard(target, "cureHeavyWounds")
			target.HitPoints -= 1
			return fmt.Sprintf("%s caused heavy wounds on %s but they were protected and only sustained minimal damage", wizard.Name, target.Selector), nil
		} else {
			target.HitPoints -= 3
			return fmt.Sprintf("%s caused heavy wounds on %s", wizard.Name, target.Selector), nil
		}
	}

	return "", nil
}
func GetCauseHeavyWoundsSpell(s *wavinghands.Spell, e error) (*CauseHeavyWounds, error) {
	if e != nil {
		return &CauseHeavyWounds{}, e
	}

	return &CauseHeavyWounds{
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
