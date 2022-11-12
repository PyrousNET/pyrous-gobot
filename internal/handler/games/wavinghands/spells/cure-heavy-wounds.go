package spells

import (
	"fmt"
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

func (cHW CureHeavyWounds) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	var returnString string = ""
	if strings.HasSuffix(wizard.Right.Sequence, cHW.Sequence) {
		wards := strings.Split(target.Wards, ",")
		wards = append(wards, "cureHeavyWounds") // Lasts one round
		target.Wards = strings.Join(wards, ",")

		returnString += fmt.Sprintf("%s has cast Cure Heavy Wounds on %s", wizard.Name, target.Selector)
	}
	if strings.HasSuffix(wizard.Left.Sequence, cHW.Sequence) {
		if returnString != "" {
			returnString += "\n"
		}
		wards := strings.Split(target.Wards, ",")
		wards = append(wards, "cureHeavyWounds") // Lasts one round
		target.Wards = strings.Join(wards, ",")

		returnString += fmt.Sprintf("%s has cast Cure Heavy Wounds on %s", wizard.Name, target.Selector)
	}

	return returnString, nil
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

func (cHW CureHeavyWounds) Clear(target *wavinghands.Living) error {
	wards := strings.Split(target.Wards, ",")
	idx := slices.Index(wards, "cureHeavyWounds")
	if idx >= 0 {
		wavinghands.Remove(wards, idx)
	}
	if len(wards) > 0 {
		target.Wards = strings.Join(wards, ",")
	} else {
		target.Wards = ""
	}
	return nil
}
