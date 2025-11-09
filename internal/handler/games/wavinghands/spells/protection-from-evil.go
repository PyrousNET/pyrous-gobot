package spells

import (
    "fmt"
    "github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
    "strings"
)

type ProtectionFromEvil struct {
    Name        string `json:"name"`
    Sequence    string `json:"sequence"`
    ShSequence  string `json:"sh-sequence"`
    Description string `json:"description"`
    Usage       string `json:"usage"`
    Damage      int    `json:"damage"`
    Resistences string `json:"resistences"`
    Protections string `json:"protections"`
}

func (pfe ProtectionFromEvil) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
    if (len(wizard.Right.Sequence) >= len(pfe.Sequence) && strings.HasSuffix(wizard.Right.Sequence, pfe.Sequence)) ||
        (len(wizard.Left.Sequence) >= len(pfe.Sequence) && strings.HasSuffix(wizard.Left.Sequence, pfe.ShSequence)) {
        wavinghands.AddTimedWard(target, "protection-from-evil", 4)
        return fmt.Sprintf("%s is surrounded by a protection from evil", target.Selector), nil
    }
    return "", nil
}

func GetProtectionFromEvilSpell(s *wavinghands.Spell, e error) (*ProtectionFromEvil, error) {
    if e != nil {
        return &ProtectionFromEvil{}, e
    }
    return &ProtectionFromEvil{
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
