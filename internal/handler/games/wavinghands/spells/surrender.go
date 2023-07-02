package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type Surrender struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistances string `json:"resistances"`
	Protections string `json:"protections"`
}

func (s *Surrender) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	if strings.HasSuffix(wizard.Right.Sequence, s.Sequence) && strings.HasSuffix(wizard.Left.Sequence, s.ShSequence) {
		wizard.Living.HitPoints = 0

		return fmt.Sprintf("%s has surrendered", wizard.Name), nil
	}

	return "", nil
}
func GetSurrenderSpell(s *wavinghands.Spell, e error) (*Surrender, error) {
	if e != nil {
		return &Surrender{}, e
	}

	return &Surrender{
		Name:        s.Name,
		Sequence:    s.Sequence,
		ShSequence:  s.ShSequence,
		Description: s.Description,
		Usage:       s.Usage,
		Damage:      s.Damage,
		Resistances: s.Resistances,
		Protections: s.Protections,
	}, nil
}
