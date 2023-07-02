package wavinghands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	minTeams = 2
	maxTeams = 2
	PREFIX   = "wh-"
	MaxWhHp  = 15
)

type (
	Hand struct {
		Sequence string `json:"sequence"`
	}
	Spell struct {
		Name        string `json:"name"`
		Sequence    string `json:"sequence"`
		ShSequence  string `json:"sh-sequence"`
		Description string `json:"description"`
		Usage       string `json:"usage"`
		Damage      int    `json:"damage"`
		Resistances string `json:"resistances"`
		Protections string `json:"protections"`
	}
	Living struct {
		Selector  string `json:"selector"`
		HitPoints int    `json:"hp"`
		Wards     string `json:"wards"`
	}
	Monster struct {
		Type        string `json:"type"`
		Living      Living `json:"living"`
		Curses      string `json:"curses"`
		Protections string `json:"protections"`
	}
	Wizard struct {
		Right       Hand      `json:"right"`
		Left        Hand      `json:"left"`
		Target      string    `json:"target"`
		Name        string    `json:"name"`
		Living      Living    `json:"living"`
		Curses      string    `json:"curses"`
		Protections string    `json:"protections"`
		Monsters    []Monster `json:"monsters"`
	}
)

func (w *Wizard) SetTarget(t string) {
	w.Target = t
}
func (w *Wizard) GetTarget() string {
	return w.Target
}

func (h *Hand) Set(s string) {
	h.Sequence = s
}

func (h *Hand) Get() []byte {
	return []byte(h.Sequence)
}

func (h *Hand) GetAt(index int) byte {
	if index < 0 {
		return ' '
	}
	return h.Sequence[index]
}

func Remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func GetMaxTeams() int {
	return maxTeams
}
func GetMinTeams() int {
	return minTeams
}

func GetHelpSpells() string {
	spells := getSpells()

	var response string = "/echo The following spells are available for Waving Hands:\n```\n"

	for _, s := range spells {
		response += fmt.Sprintf("%s: %s\n\n", s.Name, s.Usage)
	}

	response += "```\n"

	return response
}

func GetHelpSpell(chSp string) string {
	var spell Spell
	spells := getSpells()

	for _, s := range spells {
		if strings.Title(chSp) == s.Name {
			spell = s
		}
	}

	var response string
	if spell.Name == "" {
		response = fmt.Sprintf("%s wasn't a spell.\n", chSp)
	} else {
		response = fmt.Sprintf("%s is defined as follows:\n```\n", spell.Name)
		response += fmt.Sprintf("Description: %s\n", spell.Description)
		response += fmt.Sprintf("Usage: %s\n", spell.Usage)
		response += "```\n"
	}

	return response
}

func getSpells() []Spell {
	jsonFile, err := os.Open("./spells.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened spells.json")
	byteValue, _ := io.ReadAll(jsonFile)

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	var spells []Spell
	json.Unmarshal(byteValue, &spells)
	return spells
}

func GetSpell(name string) (*Spell, error) {
	spells := getSpells()

	for i, cs := range spells {
		if cs.Name == name {
			return &spells[i], nil
		}
	}

	return &Spell{}, fmt.Errorf("not found")
}
