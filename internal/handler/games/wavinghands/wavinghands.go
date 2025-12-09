package wavinghands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	minTeams = 2
	maxTeams = 6
	PREFIX   = "wh-"
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
		Damage      int    `json:"damage"`
		Element     string `json:"element"`
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
		LastRight   string    `json:"last_right"`
		LastLeft    string    `json:"last_left"`
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

func (h Hand) Get() []byte {
	return []byte(h.Sequence)
}

func (h Hand) GetAt(index int) byte {
	return h.Sequence[index]
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

type rawSpell struct {
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Gestures    []string        `json:"gestures"`
	Alternates  json.RawMessage `json:"alternates"`
	Description string          `json:"description"`
	Usage       string          `json:"usage"`
	Damage      int             `json:"damage"`
	Resistances string          `json:"resistances"`
	Protections string          `json:"protections"`
}

func getSpells() []Spell {
	candidates := []string{
		"./spells.json",
		"../spells.json",
		"../../spells.json",
		"../../../spells.json",
		"../../../../spells.json",
	}

	var lastErr error
	for _, path := range candidates {
		jsonFile, err := os.Open(path)
		if err != nil {
			lastErr = err
			continue
		}

		byteValue, readErr := io.ReadAll(jsonFile)
		jsonFile.Close()
		if readErr != nil {
			panic(fmt.Errorf("failed to read spells data from %s: %w", path, readErr))
		}

		var rawSpells []rawSpell
		if err := json.Unmarshal(byteValue, &rawSpells); err != nil {
			panic(fmt.Errorf("failed to parse spells data from %s: %w", path, err))
		}

		spells := make([]Spell, 0, len(rawSpells))
		for _, rs := range rawSpells {
			sequence, shSequence := buildSequences(rs.Gestures)
			sequence = appendAlternateSequences(sequence, rs.Alternates)
			spells = append(spells, Spell{
				Name:        rs.Name,
				Sequence:    sequence,
				ShSequence:  shSequence,
				Description: rs.Description,
				Usage:       rs.Usage,
				Damage:      rs.Damage,
				Resistances: rs.Resistances,
				Protections: rs.Protections,
			})
		}
		return spells
	}

	panic(fmt.Errorf("spells.json not found in paths %v: %w", candidates, lastErr))
}

func buildSequences(gestures []string) (string, string) {
	if len(gestures) == 0 {
		return "", ""
	}

	var builder strings.Builder
	for _, g := range gestures {
		builder.WriteString(normalizeGesture(g))
	}

	sequence := builder.String()
	shSequence := ""
	if len(gestures) == 1 && strings.HasPrefix(strings.ToLower(gestures[0]), "(") {
		shSequence = sequence
	}

	return sequence, shSequence
}

func normalizeGesture(token string) string {
	trimmed := strings.TrimSpace(token)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "(") {
		lower = strings.TrimPrefix(lower, "(")
	}

	switch lower {
	case "p", "w", "s", "d", "f", "c":
		return lower
	case "stab":
		return "stab"
	default:
		return lower
	}
}

func appendAlternateSequences(sequence string, raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return sequence
	}

	appendSeq := func(base, addition string) string {
		if addition == "" {
			return base
		}
		if base == "" {
			return addition
		}
		return base + "|" + addition
	}

	var asStrings []string
	if err := json.Unmarshal(raw, &asStrings); err == nil && len(asStrings) > 0 {
		return appendSeq(sequence, encodeGestureList(asStrings))
	}

	var nested [][]string
	if err := json.Unmarshal(raw, &nested); err == nil && len(nested) > 0 {
		for _, seq := range nested {
			sequence = appendSeq(sequence, encodeGestureList(seq))
		}
		return sequence
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return appendSeq(sequence, encodeGestureList([]string{single}))
	}

	return sequence
}

func encodeGestureList(gestures []string) string {
	var builder strings.Builder
	for _, g := range gestures {
		builder.WriteString(normalizeGesture(g))
	}
	return builder.String()
}

func GetSpell(name string) (*Spell, error) {
	spells := getSpells()

	for i, cs := range spells {
		if cs.Name == name {
			return &spells[i], nil
		}
	}

	return &Spell{}, fmt.Errorf("spell %s not found", name)
}

// CleanupWards removes expired ward effects from a Living entity
func CleanupWards(living *Living) {
	wards := tokenizeWards(living.Wards)
	if len(wards) == 0 {
		return
	}

	var keep []string
	for _, ward := range wards {
		base, duration, hasDuration := parseTimedWard(ward)
		switch {
		case base == "protection-from-evil" && hasDuration:
			if duration > 1 {
				keep = append(keep, fmt.Sprintf("%s:%d", base, duration-1))
			}
		case persistentWards[base]:
			keep = append(keep, ward)
		}
	}
	living.Wards = strings.Join(keep, ",")
}

// CleanupAllWards removes expired ward effects from all players
func CleanupAllWards(players []Wizard) {
	for i := range players {
		CleanupWards(&players[i].Living)
		// Also cleanup monster wards if needed
		for j := range players[i].Monsters {
			CleanupWards(&players[i].Monsters[j].Living)
		}
	}
}

func CleanupDeadMonsters(players []Wizard) {
	for i := range players {
		alive := players[i].Monsters[:0]
		for _, m := range players[i].Monsters {
			if m.Living.HitPoints > 0 {
				alive = append(alive, m)
			}
		}
		players[i].Monsters = alive
	}
}

var persistentWards = map[string]bool{
	"amnesia":     true,
	"anti-spell":  true,
	"resist-heat": true,
	"resist-cold": true,
}

var wardLabels = map[string]string{
	"shield":               "Shield",
	"counter-spell":        "Counter Spell",
	"magic-mirror":         "Magic Mirror",
	"resist-heat":          "Resist Heat",
	"resist-cold":          "Resist Cold",
	"protection-from-evil": "Protection from Evil",
	"anti-spell":           "Anti-Spell",
	"amnesia":              "Amnesia",
}

func FormatWards(l Living) string {
	wards := tokenizeWards(l.Wards)
	if len(wards) == 0 {
		return ""
	}
	names := make([]string, 0, len(wards))
	for _, w := range wards {
		base, duration, hasDuration := parseTimedWard(w)
		label := wardLabels[base]
		if label == "" {
			label = base
		}
		if hasDuration {
			label = fmt.Sprintf("%s (%d)", label, duration)
		}
		names = append(names, label)
	}
	return strings.Join(names, ", ")
}

func wardBase(token string) string {
	if idx := strings.Index(token, ":"); idx >= 0 {
		return token[:idx]
	}
	return token
}

func parseTimedWard(token string) (string, int, bool) {
	if idx := strings.Index(token, ":"); idx >= 0 {
		base := token[:idx]
		value := token[idx+1:]
		if d, err := strconv.Atoi(value); err == nil {
			return base, d, true
		}
		return base, 0, false
	}
	return token, 0, false
}

type MonsterStat struct {
	Damage    int
	HitPoints int
	Element   string
}

var monsterStats = map[string]MonsterStat{
	"goblin":         {Damage: 1, HitPoints: 1},
	"ogre":           {Damage: 2, HitPoints: 2},
	"troll":          {Damage: 3, HitPoints: 3},
	"giant":          {Damage: 4, HitPoints: 4},
	"fire-elemental": {Damage: 3, HitPoints: 3, Element: "fire"},
	"ice-elemental":  {Damage: 3, HitPoints: 3, Element: "cold"},
}

func GetMonsterStats(monsterType string) (MonsterStat, bool) {
	stat, ok := monsterStats[monsterType]
	return stat, ok
}

func AddMonster(wizard *Wizard, monsterType string) (Monster, error) {
	stats, ok := GetMonsterStats(monsterType)
	if !ok {
		return Monster{}, fmt.Errorf("unknown monster type %s", monsterType)
	}

	selector := fmt.Sprintf("%s:%s#%d", wizard.Name, monsterType, len(wizard.Monsters)+1)
	monster := Monster{
		Type:    monsterType,
		Damage:  stats.Damage,
		Element: stats.Element,
		Living: Living{
			Selector:  selector,
			HitPoints: stats.HitPoints,
		},
	}
	wizard.Monsters = append(wizard.Monsters, monster)
	return monster, nil
}

func tokenizeWards(raw string) []string {
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	results := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		results = append(results, p)
	}
	return results
}

// AddWard ensures the target has the named ward exactly once.
func AddWard(living *Living, ward string) {
	if ward == "" {
		return
	}
	wards := tokenizeWards(living.Wards)
	for _, w := range wards {
		if wardBase(w) == ward {
			return
		}
	}
	wards = append(wards, ward)
	living.Wards = strings.Join(wards, ",")
}

func AddTimedWard(living *Living, ward string, duration int) {
	if ward == "" || duration <= 0 {
		return
	}
	RemoveWard(living, ward)
	token := fmt.Sprintf("%s:%d", ward, duration)
	wards := tokenizeWards(living.Wards)
	wards = append(wards, token)
	living.Wards = strings.Join(wards, ",")
}

// HasWard checks whether a ward is active on the target.
func HasWard(living *Living, ward string) bool {
	if ward == "" {
		return false
	}
	wards := tokenizeWards(living.Wards)
	for _, w := range wards {
		if wardBase(w) == ward {
			return true
		}
	}
	return false
}

// RemoveWard removes a ward immediately (used after it takes effect).
func RemoveWard(living *Living, ward string) {
	if ward == "" || living.Wards == "" {
		return
	}
	wards := tokenizeWards(living.Wards)
	filtered := wards[:0]
	for _, w := range wards {
		if wardBase(w) == ward {
			continue
		}
		filtered = append(filtered, w)
	}
	living.Wards = strings.Join(filtered, ",")
}

// CounterSpellBlocks reports whether an active counter-spell nullifies an incoming spell.
func CounterSpellBlocks(target *Living, caster, spellName string) (bool, string) {
	if HasWard(target, "counter-spell") {
		return true, fmt.Sprintf("%s tried to cast %s on %s but it was countered.", caster, spellName, target.Selector)
	}
	return false, ""
}

// HasShield returns true when shield-like protection (shield, counter-spell, protection from evil) is active.
func HasShield(target *Living) bool {
	return HasWard(target, "shield") || HasWard(target, "counter-spell") || HasWard(target, "protection-from-evil")
}
