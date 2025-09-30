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
	rightHandMatch := len(wizard.Right.Sequence) >= len(cHW.Sequence) && strings.HasSuffix(wizard.Right.Sequence, cHW.Sequence)
	leftHandMatch := cHW.ShSequence != "" && len(wizard.Left.Sequence) >= len(cHW.ShSequence) && strings.HasSuffix(wizard.Left.Sequence, cHW.ShSequence)
	
	if rightHandMatch || leftHandMatch {
		wards := strings.Split(target.Wards, ",")
		wards = append(wards, "cureHeavyWounds") // Lasts one round
		target.Wards = strings.Join(wards, ",")

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
	wards := strings.Split(target.Wards, ",")
	idx := slices.Index(wards, "cureHeavyWounds")
	wavinghands.Remove(wards, idx)
	target.Wards = strings.Join(wards, ",")
	return nil
}
