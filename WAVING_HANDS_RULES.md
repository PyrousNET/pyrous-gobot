# Waving Hands - Turn-Based Wizard Dueling Game Rules

## Overview
Waving Hands is a strategic turn-based game where wizards duel by casting spells using hand gestures. Each wizard has two hands and can make different gestures to build up spell sequences.

## Basic Gameplay

### Starting a Game
- Use `wh` command to join a game
- Use `wh start` to begin the game
- Minimum 2 players, maximum 6 players

### Hand Gestures
Each turn, you must provide gestures for both hands:
- `f` - Wiggling Fingers
- `p` - Proffered Palm
- `s` - Snap
- `w` - Wave
- `d` - Digit Point
- `c` - Clap (requires both hands)
- `stab` - Stab with knife (one hand only)
- `nothing` - Do nothing (represented as `0`)

### Spell Casting
Spells are cast by making specific sequences of gestures. When your gesture sequence ends with a spell's required sequence, the spell is cast.

## Available Spells

### Basic Damage Spells
- **Cause Light Wounds** (`wpf`): Deals 2 damage
- **Cause Heavy Wounds** (`wpfd`): Deals 3 damage
- **Missile** (`sd`): Deals 1 damage, blocked by shield
- **Finger of Death** (`pwpfsssd`): Instant death, very powerful

### Healing Spells
- **Cure Light Wounds** (`dfp`): Heals 2 damage, protects against light wounds
- **Cure Heavy Wounds** (`dfpw`): Heals 2 damage, protects against heavy wounds

### Protection Spells
- **Shield** (`p`): Blocks missiles, stabs, and monster attacks for one turn
- **Counter Spell** (`wws` or `wpp`): Blocks all spells cast at target that turn

### Mental Effects
- **Anti-Spell** (`spf`): Forces target to restart spell sequences next turn
- **Amnesia** (`ddp`): Forces target to repeat previous turn's gestures

### Summoning
- **Elemental** (`cswws` both hands): Summons fire or ice elemental
- **Surrender** (`p` both hands): Surrender the game

### Physical Attacks
- **Stab** (`stab`): 1 damage, blocked by shield, can't be countered

## Game Mechanics

### Health Points
- All wizards start with 15 hit points
- Game ends when a wizard reaches 0 hit points
- Last wizard standing wins

### Ward System
- Protective spells add temporary "wards" that last one round
- Wards protect against specific types of attacks
- Multiple wards can be active simultaneously

### Turn Resolution Order
1. All players submit gestures simultaneously
2. Surrender spells are processed first
3. Protection spells are cast
4. Mental effect spells are applied
5. Damage spells are resolved
6. Physical attacks (stab) are processed
7. Summoning spells are cast
8. Round counter increases

## Strategy Tips

1. **Plan Ahead**: Many powerful spells require long sequences
2. **Timing**: Protection spells must be cast the same turn as incoming damage
3. **Misdirection**: Start multiple spell sequences to confuse opponents
4. **Resource Management**: Balance offense and defense
5. **Quick Attacks**: Missile and stab are fast but weak
6. **Counter-play**: Use anti-spell to disrupt long casting sequences

## Help Commands
- `wh help-spells`: List all available spells
- `wh help <spell-name>`: Get detailed information about a specific spell

## Example Turn
```
# Turn 1: Start building up to Finger of Death
Your gestures: p w

# Turn 2: Continue sequence
Your gestures: p f

# Turn 3: Continue sequence  
Your gestures: s s

# Turn 4: Complete the spell
Your gestures: s d target-wizard
# Finger of Death is cast at target-wizard!
```

Remember: Waving Hands is a game of strategy, timing, and misdirection. May the best wizard win!