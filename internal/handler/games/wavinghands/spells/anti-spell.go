package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
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
	if blocked, msg := wavinghands.CounterSpellBlocks(target, wizard.Name, as.Name); blocked {
		return msg, nil
	}

	rightMatch := len(wizard.Right.Sequence) >= len(as.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, as.Sequence)
	leftPattern := as.Sequence
	if as.ShSequence != "" {
		leftPattern = as.ShSequence
	}
	leftMatch := len(wizard.Left.Sequence) >= len(leftPattern) &&
		strings.HasSuffix(wizard.Left.Sequence, leftPattern)

	if rightMatch || leftMatch {

		wavinghands.AddWard(target, "anti-spell")

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
	wavinghands.RemoveWard(target, "anti-spell")
	return nil
}
