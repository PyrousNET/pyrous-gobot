package games

import "strings"

func (h BotGameHelp) Rps(request BotGame) (response HelpResponse) {
	response.Description = "Play Rock Paper Scissors with another player."
	response.Help = strings.TrimSpace(`
Play Rock Paper Scissors with another player.

Usage:
  $rps [in <channel>] <rock|paper|scissors>
  $rps <channel> <rock|paper|scissors>

Notes:
- First player queues a game in the channel; the next player in the same game resolves the match.
- Choices are DM'd to each player and results are echoed in the channel.`)

	return response
}

func (h BotGameHelp) Wh(request BotGame) (response HelpResponse) {
	response.Description = "Play Waving Hands (turn-based wizard duels)."
	response.Help = strings.TrimSpace(`
Play Waving Hands (turn-based wizard duels).

Commands:
  $wh [in <channel>]           - Join the lobby in that channel
  $wh start                    - Start when 2-6 players have joined
  $wh help-spells              - List available spells and gestures
  $wh help <spell>             - Show detailed help for a spell
  $wh <channel> <R> <L> [target] - Submit gestures (right, left, optional target)

See WAVING_HANDS_RULES.md for complete rules and examples.`)

	return response
}

func (h BotGameHelp) Farkle(request BotGame) (response HelpResponse) {
	response.Description = "Farkle (dice game) — coming soon."
	response.Help = strings.TrimSpace(`
Farkle (dice game) — not implemented yet, planned for a future release.`)

	return response
}
