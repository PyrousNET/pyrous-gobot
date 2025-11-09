package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type FingerOfDeath struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (fod FingerOfDeath) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	rightMatch := len(wizard.Right.Sequence) >= len(fod.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, fod.Sequence)
	leftPattern := fod.Sequence
	if fod.ShSequence != "" {
		leftPattern = fod.ShSequence
	}
	leftMatch := len(wizard.Left.Sequence) >= len(leftPattern) &&
		strings.HasSuffix(wizard.Left.Sequence, leftPattern)

	if rightMatch || leftMatch {

		// Finger of Death is unaffected by counter-spell but can be stopped by dispel magic
		// For simplicity, we'll implement it as instant death
		target.HitPoints = 0
		return fmt.Sprintf("%s has cast Finger of Death on %s - they are killed instantly!", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetFingerOfDeathSpell(s *wavinghands.Spell, e error) (*FingerOfDeath, error) {
	if e != nil {
		return &FingerOfDeath{}, e
	}

	return &FingerOfDeath{
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
