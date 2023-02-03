package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
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
	var returnString string = ""
	if strings.HasSuffix(wizard.Right.Sequence, "1") {
		wizard.Right.Set(strings.TrimRight(wizard.Right.Sequence, "1"))

		if strings.Contains(target.Wards, "shield") {
			returnString += fmt.Sprintf("%s tries to stab %s with right hand but they were protected by a shield and took no damage", wizard.Name, target.Selector)
		} else {
			target.HitPoints -= s.Damage
			returnString += fmt.Sprintf("%s stabbed %s with right hand", wizard.Name, target.Selector)
		}
	}
	if strings.HasSuffix(wizard.Left.Sequence, "1") {
		wizard.Left.Set(strings.TrimRight(wizard.Left.Sequence, "1"))
		if returnString != "" {
			returnString += "\n"
		}
		if strings.Contains(target.Wards, "shield") {
			returnString += fmt.Sprintf("%s tries to stab %s with left hand but they were protected by a shield and took no damage", wizard.Name, target.Selector)
		} else {
			target.HitPoints -= s.Damage
			returnString += fmt.Sprintf("%s stabbed %s with left hand", wizard.Name, target.Selector)
		}
	}

	return returnString, nil
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
