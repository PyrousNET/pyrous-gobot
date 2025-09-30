package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"golang.org/x/exp/slices"
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
		
		if target.Wards == "" {
			target.Wards = "cureLightWounds"
		} else {
			target.Wards = target.Wards + ",cureLightWounds"
		}

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
	wards := strings.Split(target.Wards, ",")
	idx := slices.Index(wards, "cureLightWounds")
	if idx >= 0 {
		wards = wavinghands.Remove(wards, idx)
		target.Wards = strings.Join(wards, ",")
	}
	return nil
}