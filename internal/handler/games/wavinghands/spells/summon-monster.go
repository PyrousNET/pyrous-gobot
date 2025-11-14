package spells

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"strings"
)

type SummonMonster struct {
	Name        string `json:"name"`
	Sequence    string `json:"sequence"`
	ShSequence  string `json:"sh-sequence"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Damage      int    `json:"damage"`
	Resistences string `json:"resistences"`
	Protections string `json:"protections"`
	MonsterType string
}

func (sm SummonMonster) Cast(wizard *wavinghands.Wizard, target *wavinghands.Living) (string, error) {
	rightMatch := len(wizard.Right.Sequence) >= len(sm.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, sm.Sequence)
	leftMatch := len(wizard.Left.Sequence) >= len(sm.Sequence) &&
		strings.HasSuffix(wizard.Left.Sequence, sm.Sequence)

	if rightMatch || leftMatch {
		monster, err := wavinghands.AddMonster(wizard, sm.MonsterType)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s summons a %s", wizard.Name, monster.Type), nil
	}

	return "", nil
}

func getSummonMonsterSpell(s *wavinghands.Spell, e error, monsterType string) (*SummonMonster, error) {
	if e != nil {
		return &SummonMonster{}, e
	}
	return &SummonMonster{
		Name:        s.Name,
		Sequence:    s.Sequence,
		ShSequence:  s.ShSequence,
		Description: s.Description,
		Usage:       s.Usage,
		Damage:      s.Damage,
		Resistences: s.Resistances,
		Protections: s.Protections,
		MonsterType: monsterType,
	}, nil
}

func GetSummonGoblinSpell(s *wavinghands.Spell, e error) (*SummonMonster, error) {
	return getSummonMonsterSpell(s, e, "goblin")
}

func GetSummonOgreSpell(s *wavinghands.Spell, e error) (*SummonMonster, error) {
	return getSummonMonsterSpell(s, e, "ogre")
}

func GetSummonTrollSpell(s *wavinghands.Spell, e error) (*SummonMonster, error) {
	return getSummonMonsterSpell(s, e, "troll")
}

func GetSummonGiantSpell(s *wavinghands.Spell, e error) (*SummonMonster, error) {
	return getSummonMonsterSpell(s, e, "giant")
}
