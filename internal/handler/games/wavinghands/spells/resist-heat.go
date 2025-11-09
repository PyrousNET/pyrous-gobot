package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type ResistHeat struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (rh ResistHeat) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if (len(wizard.Right.Sequence) >= len(rh.Sequence) && strings.HasSuffix(wizard.Right.Sequence, rh.Sequence)) ||
		(len(wizard.Left.Sequence) >= len(rh.Sequence) && strings.HasSuffix(wizard.Left.Sequence, rh.ShSequence)) {
		wavinghands.AddWard(target, "resist-heat")
		return fmt.Sprintf("%s is now resistant to heat", target.Selector), nil
	}

	return "", nil
}

func GetResistHeatSpell(s *wavinghands.Spell, e error) (*ResistHeat, error) {
	if e != nil {
		return &ResistHeat{}, e
	}

	return &ResistHeat{
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
