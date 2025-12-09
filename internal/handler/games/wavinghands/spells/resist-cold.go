package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type ResistCold struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (rc ResistCold) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	rightMatch := len(wizard.Right.Sequence) >= len(rc.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, rc.Sequence)
	leftMatch := len(wizard.Left.Sequence) >= len(rc.Sequence) &&
		strings.HasSuffix(wizard.Left.Sequence, rc.Sequence)

	if rightMatch || leftMatch {
		wavinghands.AddWard(target, "resist-cold")
		return fmt.Sprintf("%s is now resistant to cold", target.Selector), nil
	}

	return "", nil
}

func GetResistColdSpell(s *wavinghands.Spell, e error) (*ResistCold, error) {
	if e != nil {
		return &ResistCold{}, e
	}

	return &ResistCold{
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
