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
	maxTeams = 6
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
		Right       Hand    `json:"right"`
		Left        Hand    `json:"left"`
		Name        string  `json:"name"`
		Living      Living  `json:"living"`
		Curses      string  `json:"curses"`
		Protections string  `json:"protections"`
		Monsters    Monster `json:"monsters"`
	}
)

func (h *Hand) Set(s string) {
	h.Sequence = s
}

func (h Hand) Get() []byte {
	return []byte(h.Sequence)
}

func (h Hand) GetAt(index int) byte {
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
		response = fmt.Sprintf("/echo %s wasn't a spell.\n", chSp)
	} else {
		response = fmt.Sprintf("/echo %s is defined as follows:\n```\n", spell.Name)
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
