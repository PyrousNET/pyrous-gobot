# Waving Hands Reference

> Adapted from the original rules published by Richard Bartle on
> [gamecabinet.com](http://www.gamecabinet.com/rules/WavingHands.html).
> The notes below describe how the Mattermost bot currently behaves and
> where it still differs from the full tabletop game.

## Getting Started in Mattermost

- Use `wh` in a channel to join/ create a lobby (2–6 players).
- Run `wh start` once everyone has joined.
- Submit turns with `wh <channel> <right-gesture> <left-gesture> [target]`.
  - Gestures are single letters: `f` (wiggle fingers), `p` (proffered palm),
    `s` (snap), `w` (wave), `d` (digit point), `c` (clap, requires both
    hands), `stab`, or `nothing`.
  - Targets default to yourself but can be `opponent`, `name`, or
    `name:monster`.
- Helpful commands:
  - `wh status` – show HP and current protections.
  - `wh help-spells` – list every spell the bot knows.
  - `wh help <spell>` – details for a single spell.
  - `wh rules` – short reminder plus link to this file.

## Turn Structure

1. **Plan gestures** (one per hand) secretly and submit via the bot.
2. **Reveal**: after every player has provided gestures, the bot
   announces what everyone just did (unless hidden by a future
   invisibility effect).
3. **Resolve spells** in batches so simultaneous effects make sense:

   1. Surrender (both hands `p`) – the wizard drops to 0 HP immediately.
   2. Defenses – Shield, Counter‑Spell, Cure Light/Heavy Wounds.
   3. Mental effects – Anti‑Spell, Amnesia (future versions will add the
      rest of the official enchantments).
   4. Damage – Finger of Death, Cause Wounds, Missile, Stab.
   5. Summons – Elemental (other monsters still to come).

4. Apply monster attacks, cleanup wards that only last one round, and
   advance the round counter. The last wizard standing wins.

## Gestures and Casting

- Gestures chain together across rounds. Whenever a hand’s current
  sequence ends with a spell pattern from `spells.json` the spell fires.
- Gestures can contribute to multiple spells simultaneously. Only one
  spell may be emitted per gesture; if two spells would complete on the
  same final gesture, the caster chooses which effect to resolve.
- Non-gestures (`stab`, `nothing`, or a failed two-handed gesture such as
  a single `c`) break every active sequence for that hand.
- Multi-handed requirements (e.g., claps or the final `((w` of
  invisibility in the tabletop rules) must be entered for both hands in
  the same turn when the bot eventually supports those spells.

## Implemented Spells

| Category        | Spell (gestures)                                  | Notes                                                                            |
|-----------------|---------------------------------------------------|----------------------------------------------------------------------------------|
| Protection      | Shield (`p`), Counter‑Spell (`wws` or `wpp`), Remove Enchantment (`pdwp`), Magic Mirror (`c(w`), Dispel Magic (`cdpw`) | Magic Mirror reflects single-target spells for a turn; Dispel wipes all wards/monsters and blocks every spell that turn. |
| Resistances     | Resist Heat (`wwfp`), Resist Cold (`ssfp`)         | Adds permanent wards that flag the target as heat/cold immune for future effects.|
| Healing         | Cure Light (`dfp`), Cure Heavy (`dfpw`)            | Adds short‑lived wards that blunt incoming Cause Wounds of the same weight.      |
| Mental          | Anti‑Spell (`spf`), Amnesia (`ddp`)                | Anti‑Spell forces the target to restart sequences next turn; Amnesia repeats gestures. |
| Damage          | Missile (`sd`), Cause Light (`wpf`), Cause Heavy (`wpfd`), Finger of Death (`pwpfsssd`) | Missile/Stab respect Shield and counter effects. Finger of Death ignores Counter. |
| Physical        | Stab                                              | Simple 1 HP attack; cannot be reflected.                                         |
| Summoning       | Elemental (`cswws` with both hands), Goblin (`sfw`), Ogre (`psfw`), Troll (`fpsfw`), Giant (`wfsfw`) | Summoned monsters stick around, have their own HP, and swing at the controller’s chosen target each round. |
| Utility         | Surrender (`p` + `p`)                             | Immediately removes the caster from the duel.                                    |

## Not Yet Implemented (Coming Soon)

The original rules describe additional mechanics that are **not** in the
bot yet:

- Remaining protection spells (Protection from Evil).
- Offensive area spells (Fireball, Lightning Bolt, Fire/Ice Storm),
  delayed spells, and poison/disease timers.
- Enchantments such as Charm Person/Monster, Confusion, Fear, Paralysis,
  Blindness, Haste, Time Stop, Delayed Effect, Permanency.
- Visibility tricks (Invisibility).

Until those features land, play within the spell list above for
consistent behavior.

## Key Rule Clarifications

- **Health**: Wizards begin at 15 HP. Reaching 0 HP eliminates the
  player. Healing cannot push a wizard above 15.
- **Counter‑Spell**: Protects against every hostile spell (except Dispel
  Magic and Finger of Death per the published rules) for the rest of the
  round. The bot now models this by short-lived wards that other spells
  respect.
- **Anti‑Spell**: Forces the target to abandon all in‑progress spell
  chains after the current round. On the next submission the bot resets
  both hands’ gesture history before appending the new gestures.
- **Amnesia**: The target must repeat the exact gestures they made this
  round when they submit their next turn. The bot enforces this
  automatically and informs affected players.
- **Wards**: The bot tracks protections on each living creature. Wards
  expire automatically unless otherwise noted. Protection from Evil
  lingers for four turns, Resist Heat/Cold persist until a Remove
  Enchantment or Dispel Magic clears them. You can see active wards via
  `wh status`.
- **Magic Mirror**: Reflects single-target spells back at the caster
  unless the spell was self-targeted or an area effect. Counter-Spell
  or Dispel Magic on the same turn cancels the mirror.
- **Dispel Magic**: When it resolves, no other spells work that round,
  every ward is cleared, and all monsters are banished after their
  current attack. It also grants the caster a shield for that turn.

## Strategy Reminders

- Keep multiple spell chains alive so you can pivot between offense and
  defense.
- Shield + Counter‑Spell forces opponents to either disrupt you or
  commit to high-cost spells like Finger of Death.
- Anti‑Spell and Amnesia are tempo plays—use them when you suspect an
  opponent is close to completing a big sequence.
- Stabs are fast but predictable. Mix them with missile pressure to
  draw out shields before you cast larger spells.

## Attribution

These rules are a paraphrased digest of Richard Bartle’s *Waving Hands*
design (originally distributed as Spellbinder). See the link at the top
of this document for the complete text and advanced options not yet
ported to the bot.
