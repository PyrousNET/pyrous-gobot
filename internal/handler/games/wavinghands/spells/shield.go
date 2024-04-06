package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"golang.org/x/exp/slices"
	"strings"
)

type Shield struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistances string `json:"resistances"`
	Protections string `json:"protections"`
}

func (s *Shield) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	var returnString string
	if strings.HasSuffix(wizard.Right.Sequence, s.Sequence) || strings.HasSuffix(wizard.Left.Sequence, s.Sequence) {
		wards := strings.Split(target.Wards, ",")
		wards = append(wards, "shield")
		target.Wards = strings.Join(wards, ",")

		returnString = fmt.Sprintf("%s has cast Shield on %s", wizard.Name, target.Selector)
	}
	return returnString, nil
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
		Resistances: s.Resistances,
		Protections: s.Protections,
	}, nil
}

func (s *Shield) Clear(target *wavinghands.Living) error {
	if target.Wards == "" {
		return nil
	}
	wards := strings.Split(target.Wards, ",")
	idx := slices.Index(wards, "shield")
	if idx >= 0 {
		wards = wavinghands.Remove(wards, idx)
	}
	wards = removeEmptyValues(wards)
	if len(wards) > 0 {
		target.Wards = strings.Join(wards, ",")
	} else {
		target.Wards = ""
	}
	return nil
}

func removeEmptyValues(items []string) []string {
	var result []string
	for _, str := range items {
		if str != "" {
			result = append(result, str)
		}
	}

	return result
}
