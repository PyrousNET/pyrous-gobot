package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"golang.org/x/exp/slices"
	"strings"
)

type AntiSpell struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (as AntiSpell) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if (len(wizard.Right.Sequence) >= len(as.Sequence) && strings.HasSuffix(wizard.Right.Sequence, as.Sequence)) ||
		(len(wizard.Left.Sequence) >= len(as.Sequence) && strings.HasSuffix(wizard.Left.Sequence, as.ShSequence)) {

		if target.Wards == "" {
			target.Wards = "anti-spell"
		} else {
			target.Wards = target.Wards + ",anti-spell"
		}

		return fmt.Sprintf("%s has cast Anti-Spell on %s", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetAntiSpellSpell(s *wavinghands.Spell, e error) (*AntiSpell, error) {
	if e != nil {
		return &AntiSpell{}, e
	}

	return &AntiSpell{
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

func (as AntiSpell) clear(target *wavinghands.Living) error {
	wards := strings.Split(target.Wards, ",")
	idx := slices.Index(wards, "anti-spell")
	if idx >= 0 {
		wards = wavinghands.Remove(wards, idx)
		target.Wards = strings.Join(wards, ",")
	}
	return nil
}