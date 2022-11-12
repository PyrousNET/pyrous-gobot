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
	var returnString string = ""
	if strings.HasSuffix(wizard.Right.Sequence, m.Sequence) {

		if strings.Contains(target.Wards, "shield") {
			returnString += fmt.Sprintf("%s cast a missile at %s but they were protected by a shield and took no damage", wizard.Name, target.Selector)
		} else {
			target.HitPoints -= m.Damage
			returnString += fmt.Sprintf("%s cast a missile at %s", wizard.Name, target.Selector)
		}
	}
	if strings.HasSuffix(wizard.Left.Sequence, m.Sequence) {
		if returnString != "" {
			returnString += "\n"
		}
		if strings.Contains(target.Wards, "shield") {
			returnString += fmt.Sprintf("%s cast a missile at %s but they were protected by a shield and took no damage", wizard.Name, target.Selector)
		} else {
			target.HitPoints -= m.Damage
			returnString += fmt.Sprintf("%s cast a missile at %s", wizard.Name, target.Selector)
		}
	}

	return returnString, nil
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
