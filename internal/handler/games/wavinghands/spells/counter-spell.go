package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type CounterSpell struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (cs CounterSpell) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	// Counter-spell can be cast with two different sequences: "wws" or "wpp"
	sequences := strings.Split(cs.Sequence, "|")
	canCast := false

	for _, seq := range sequences {
		if (len(wizard.Right.Sequence) >= len(seq) && strings.HasSuffix(wizard.Right.Sequence, seq)) ||
			(len(wizard.Left.Sequence) >= len(seq) && strings.HasSuffix(wizard.Left.Sequence, seq)) {
			canCast = true
			break
		}
	}

	if canCast {
		wavinghands.AddWard(target, "counter-spell")

		return fmt.Sprintf("%s has cast Counter-Spell on %s", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetCounterSpellSpell(s *wavinghands.Spell, e error) (*CounterSpell, error) {
	if e != nil {
		return &CounterSpell{}, e
	}

	return &CounterSpell{
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

func (cs CounterSpell) clear(target *wavinghands.Living) error {
	wavinghands.RemoveWard(target, "counter-spell")
	return nil
}
