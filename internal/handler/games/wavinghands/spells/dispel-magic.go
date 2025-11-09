package spells

import (
    "fmt"
    "github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
    "strings"
)

type DispelMagic struct {
    Name        string `json:"name"`
    Sequence    string `json:"sequence"`
    ShSequence  string `json:"sh-sequence"`
    Description string `json:"description"`
    Usage       string `json:"usage"`
    Damage      int    `json:"damage"`
    Resistences string `json:"resistences"`
    Protections string `json:"protections"`
}

func (dm DispelMagic) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (bool, string, error) {
    if (len(wizard.Right.Sequence) >= len(dm.Sequence) && strings.HasSuffix(wizard.Right.Sequence, dm.Sequence)) ||
        (len(wizard.Left.Sequence) >= len(dm.Sequence) && strings.HasSuffix(wizard.Left.Sequence, dm.ShSequence)) {
        return true, fmt.Sprintf("%s unleashes Dispel Magic!", wizard.Name), nil
    }
    return false, "", nil
}

func GetDispelMagicSpell(s *wavinghands.Spell, e error) (*DispelMagic, error) {
    if e != nil {
        return &DispelMagic{}, e
    }
    return &DispelMagic{
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
