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
	response.Description = "Play Farkle (dice) with scoring and final round."
	response.Help = strings.TrimSpace(`
Play Farkle (dice) with standard scoring and a final round at 5,000.

Usage:
  $farkle [in <channel>]   - Join or create a lobby
  $farkle start            - Start when 2+ players joined
  $farkle roll             - Roll your remaining dice
  $farkle keep <dice>      - Keep scoring dice from your last roll (e.g. $farkle keep 1 5 5)
  $farkle bank             - Bank turn points and pass the turn
  $farkle quit             - End/clear the current game

Scoring highlights:
- 1s=100, 5s=50; triples: 1s=1000, 2/3/4/5/6 = 200/300/400/500/600
- 4/5/6 of a kind add another triple base each die
- Straight (1-6)=1500, three pairs=1500, two triplets=2500, four-kind+pair=1500
- Hot dice: score all dice in a roll and you may roll all 6 again
- Final round triggers when a player banks >= 5000; others get one last turn to beat them.`)

	return response
}
