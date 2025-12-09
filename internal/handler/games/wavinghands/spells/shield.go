package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type Shield struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (s Shield) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	rightMatch := len(wizard.Right.Sequence) >= len(s.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, s.Sequence)
	leftPattern := s.Sequence
	if s.ShSequence != "" {
		leftPattern = s.ShSequence
	}
	leftMatch := len(wizard.Left.Sequence) >= len(leftPattern) &&
		strings.HasSuffix(wizard.Left.Sequence, leftPattern)

	if rightMatch || leftMatch {

		wavinghands.AddWard(target, "shield")

		return fmt.Sprintf("%s has cast Shield on %s", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetShieldSpell(s *wavinghands.Spell, e error) (*Shield, error) {
	if e != nil {
		return &Shield{}, e
	}

	return &Shield{
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

func (s Shield) clear(target *wavinghands.Living) error {
	wavinghands.RemoveWard(target, "shield")
	return nil
}
