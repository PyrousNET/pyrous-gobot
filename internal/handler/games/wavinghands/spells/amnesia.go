package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type Amnesia struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (a Amnesia) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if blocked, msg := wavinghands.CounterSpellBlocks(target, wizard.Name, a.Name); blocked {
		return msg, nil
	}

	rightMatch := len(wizard.Right.Sequence) >= len(a.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, a.Sequence)
	leftPattern := a.Sequence
	if a.ShSequence != "" {
		leftPattern = a.ShSequence
	}
	leftMatch := len(wizard.Left.Sequence) >= len(leftPattern) &&
		strings.HasSuffix(wizard.Left.Sequence, leftPattern)

	if rightMatch || leftMatch {

		// Check if target has conflicting enchantments (confusion, charm person, charm monster, paralysis, fear)
		conflictingSpells := []string{"confusion", "charm-person", "charm-monster", "paralysis", "fear"}
		hasConflict := false
		for _, spell := range conflictingSpells {
			if wavinghands.HasWard(target, spell) {
				hasConflict = true
				break
			}
		}

		if hasConflict {
			return fmt.Sprintf("%s cast Amnesia on %s but it had no effect due to conflicting enchantments", wizard.Name, target.Selector), nil
		}

		wavinghands.AddWard(target, "amnesia")

		return fmt.Sprintf("%s has cast Amnesia on %s", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetAmnesiaSpell(s *wavinghands.Spell, e error) (*Amnesia, error) {
	if e != nil {
		return &Amnesia{}, e
	}

	return &Amnesia{
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

func (a Amnesia) clear(target *wavinghands.Living) error {
	wavinghands.RemoveWard(target, "amnesia")
	return nil
}
