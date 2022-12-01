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
	var returnString string = ""
	if strings.HasSuffix(wizard.Right.Sequence, cHW.Sequence) {

		if strings.Contains(target.Wards, "cureHeavyWounds") {
			target.HitPoints -= cHW.Damage - 2
			returnString = fmt.Sprintf("%s caused heavy wounds on %s but they were protected and only sustained minimal damage", wizard.Name, target.Selector)
		} else {
			target.HitPoints -= cHW.Damage
			returnString = fmt.Sprintf("%s caused heavy wounds on %s", wizard.Name, target.Selector)
		}
	}
	if strings.HasSuffix(wizard.Left.Sequence, cHW.Sequence) {
		if returnString != "" {
			returnString = returnString + "\n"
		}
		if strings.Contains(target.Wards, "cureHeavyWounds") {
			target.HitPoints -= cHW.Damage - 2
			returnString += fmt.Sprintf("%s caused heavy wounds on %s but they were protected and only sustained minimal damage", wizard.Name, target.Selector)
		} else {
			target.HitPoints -= cHW.Damage
			returnString += fmt.Sprintf("%s caused heavy wounds on %s", wizard.Name, target.Selector)
		}
	}

	return returnString, nil
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
