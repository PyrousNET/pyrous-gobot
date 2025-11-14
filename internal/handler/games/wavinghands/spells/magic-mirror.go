package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type MagicMirror struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (mm MagicMirror) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	rightMatch := len(wizard.Right.Sequence) >= len(mm.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, mm.Sequence)
	leftMatch := len(wizard.Left.Sequence) >= len(mm.Sequence) &&
		strings.HasSuffix(wizard.Left.Sequence, mm.Sequence)

	if rightMatch || leftMatch {
		if wavinghands.HasWard(target, "counter-spell") {
			return "", nil
		}
		wavinghands.AddWard(target, "magic-mirror")
		return fmt.Sprintf("%s conjures a magic mirror", wizard.Name), nil
	}
	return "", nil
}

func GetMagicMirrorSpell(s *wavinghands.Spell, e error) (*MagicMirror, error) {
	if e != nil {
		return &MagicMirror{}, e
	}
	return &MagicMirror{
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
