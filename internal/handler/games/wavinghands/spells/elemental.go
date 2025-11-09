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
	if len(wizard.Right.Sequence) >= len(e.Sequence) && len(wizard.Left.Sequence) >= len(e.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, e.Sequence) &&
		strings.HasSuffix(wizard.Left.Sequence, e.Sequence) {

		monster, err := wavinghands.AddMonster(wizard, "fire-elemental")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s summons a %s", wizard.Name, monster.Type), nil
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
