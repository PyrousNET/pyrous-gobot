package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type RemoveEnchantment struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
}

func (re RemoveEnchantment) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	rightMatch := len(wizard.Right.Sequence) >= len(re.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, re.Sequence)
	leftMatch := len(wizard.Left.Sequence) >= len(re.Sequence) &&
		strings.HasSuffix(wizard.Left.Sequence, re.Sequence)

	if rightMatch || leftMatch {
		target.Wards = ""
		return fmt.Sprintf("%s has removed all enchantments from %s", wizard.Name, target.Selector), nil
	}

	return "", nil
}

func GetRemoveEnchantmentSpell(s *wavinghands.Spell, e error) (*RemoveEnchantment, error) {
	if e != nil {
		return &RemoveEnchantment{}, e
	}

	return &RemoveEnchantment{
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
