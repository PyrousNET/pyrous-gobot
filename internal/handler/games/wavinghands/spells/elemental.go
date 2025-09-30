package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type Elemental struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (e Elemental) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	// Elemental requires both hands: "cswws"
	if len(wizard.Right.Sequence) >= 5 && len(wizard.Left.Sequence) >= 5 &&
		strings.HasSuffix(wizard.Right.Sequence, "cswws") &&
		strings.HasSuffix(wizard.Left.Sequence, "cswws") {
		
		// The caster decides the type after seeing all gestures (fire or ice)
		// For simplicity, we'll alternate or use a simple rule
		elementalType := "fire" // Could be made more dynamic
		
		// Add the elemental as a monster to the target wizard
		// This is a simplification - in the full game, elementals would be separate entities
		if target.Wards == "" {
			target.Wards = fmt.Sprintf("%s-elemental", elementalType)
		} else {
			target.Wards = target.Wards + fmt.Sprintf(",%s-elemental", elementalType)
		}

		return fmt.Sprintf("%s has summoned a %s elemental targeting %s", wizard.Name, elementalType, target.Selector), nil
	}

	return "", nil
}

func GetElementalSpell(s *wavinghands.Spell, e error) (*Elemental, error) {
	if e != nil {
		return &Elemental{}, e
	}

	return &Elemental{
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