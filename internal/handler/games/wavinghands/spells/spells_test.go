package spells

import (
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"testing"
)

func TestCauseLightWounds_Cast(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "wpf"},
		Left:  wavinghands.Hand{Sequence: ""},
		Name:  "TestWizard",
	}
	target := &wavinghands.Living{
		Selector:  "target",
		HitPoints: 15,
		Wards:     "",
	}

	spell := CauseLightWounds{
		Sequence: "wpf",
	}

	result, err := spell.Cast(wizard, target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if target.HitPoints != 13 { // 15 - 2 = 13
		t.Errorf("Expected HitPoints to be 13, got %d", target.HitPoints)
	}
	if result == "" {
		t.Errorf("Expected result message, got empty string")
	}
}

func TestMissile_Cast(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "sd"},
		Left:  wavinghands.Hand{Sequence: ""},
		Name:  "TestWizard",
	}
	target := &wavinghands.Living{
		Selector:  "target",
		HitPoints: 15,
		Wards:     "",
	}

	spell := Missile{
		Sequence: "sd",
	}

	result, err := spell.Cast(wizard, target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if target.HitPoints != 14 { // 15 - 1 = 14
		t.Errorf("Expected HitPoints to be 14, got %d", target.HitPoints)
	}
	if result == "" {
		t.Errorf("Expected result message, got empty string")
	}
}

func TestShield_Cast(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "p"},
		Left:  wavinghands.Hand{Sequence: ""},
		Name:  "TestWizard",
	}
	target := &wavinghands.Living{
		Selector:  "target",
		HitPoints: 15,
		Wards:     "",
	}

	spell := Shield{
		Sequence: "p",
	}

	result, err := spell.Cast(wizard, target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if target.Wards != "shield" {
		t.Errorf("Expected Wards to contain 'shield', got '%s'", target.Wards)
	}
	if result == "" {
		t.Errorf("Expected result message, got empty string")
	}
}

func TestMissile_BlockedByShield(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "sd"},
		Left:  wavinghands.Hand{Sequence: ""},
		Name:  "TestWizard",
	}
	target := &wavinghands.Living{
		Selector:  "target",
		HitPoints: 15,
		Wards:     "shield",
	}

	spell := Missile{
		Sequence: "sd",
	}

	result, err := spell.Cast(wizard, target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if target.HitPoints != 15 { // Should remain 15 due to shield
		t.Errorf("Expected HitPoints to remain 15, got %d", target.HitPoints)
	}
	if result == "" {
		t.Errorf("Expected result message about shield blocking, got empty string")
	}
}

func TestFingerOfDeath_Cast(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "pwpfsssd"},
		Left:  wavinghands.Hand{Sequence: ""},
		Name:  "TestWizard",
	}
	target := &wavinghands.Living{
		Selector:  "target",
		HitPoints: 15,
		Wards:     "",
	}

	spell := FingerOfDeath{
		Sequence: "pwpfsssd",
	}

	result, err := spell.Cast(wizard, target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if target.HitPoints != 0 { // Should be killed instantly
		t.Errorf("Expected HitPoints to be 0 (dead), got %d", target.HitPoints)
	}
	if result == "" {
		t.Errorf("Expected result message about instant death, got empty string")
	}
}

func TestStab_Cast(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "1"}, // "1" represents stab
		Left:  wavinghands.Hand{Sequence: ""},
		Name:  "TestWizard",
	}
	target := &wavinghands.Living{
		Selector:  "target",
		HitPoints: 15,
		Wards:     "",
	}

	spell := Stab{
		Sequence: "1",
	}

	result, err := spell.Cast(wizard, target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if target.HitPoints != 14 { // 15 - 1 = 14
		t.Errorf("Expected HitPoints to be 14, got %d", target.HitPoints)
	}
	if result == "" {
		t.Errorf("Expected result message about stab, got empty string")
	}
}

func TestCounterSpell_Cast(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "wpp"},
		Left:  wavinghands.Hand{Sequence: ""},
		Name:  "TestWizard",
	}
	target := &wavinghands.Living{
		Selector:  "target",
		HitPoints: 15,
		Wards:     "",
	}

	spell := CounterSpell{
		Sequence: "wws|wpp", // Can be cast with either sequence
	}

	result, err := spell.Cast(wizard, target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if target.Wards != "counter-spell" {
		t.Errorf("Expected Wards to contain 'counter-spell', got '%s'", target.Wards)
	}
	if result == "" {
		t.Errorf("Expected result message about counter spell, got empty string")
	}
}