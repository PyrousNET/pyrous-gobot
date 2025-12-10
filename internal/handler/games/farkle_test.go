package games

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func TestFarkleEndToEndFinalRound(t *testing.T) {
	c := cache.GetLocalCache()
	addTestUser(t, c, "alice")
	addTestUser(t, c, "bob")

	channel := &model.Channel{Id: "chan1", Name: "chan1"}
	responses := make(chan comms.Response, 20)
	play := func(sender, body string) {
		t.Helper()
		err := BotGame{}.Farkle(BotGame{
			body:            body,
			sender:          sender,
			ReplyChannel:    channel,
			ResponseChannel: responses,
			Cache:           c,
		})
		if err != nil {
			t.Fatalf("call from %s failed: %v", sender, err)
		}
	}

	// Join lobby
	play("alice", "")
	play("bob", "")
	play("alice", "start")

	// Lower the goal to finish quickly.
	game, ok, err := loadFarkle(channel.Id, c)
	if err != nil || !ok {
		t.Fatalf("expected game in cache: %v", err)
	}
	game.TargetScore = 500
	saveFarkle(game, c)

	// Deterministic rolls.
	originalRoll := rollDiceFn
	defer func() { rollDiceFn = originalRoll }()
	rollDiceFn = (&testRoller{rolls: [][]int{
		{1, 1, 1, 1, 1, 1}, // Alice
		{2, 3, 4, 6, 2, 3}, // Bob (farkle, ends final round)
	}}).roll

	play("alice", "roll")
	play("alice", "keep 1 1 1 1 1 1")
	play("alice", "bank") // triggers final round

	play("bob", "roll") // farkle in final round ends game

	// Collect final response
	close(responses)
	var lastMsg string
	for r := range responses {
		lastMsg = r.Message
	}

	if !strings.Contains(lastMsg, "Winner: alice") {
		t.Fatalf("expected final winner message for alice, got: %s", lastMsg)
	}

	// Game state should be cleared.
	_, ok, _ = c.Get(farklePrefix + channel.Id)
	if ok {
		t.Fatalf("expected farkle game to be cleared after completion")
	}
}

func TestFarkleTracksPointsWhenUserIdMissing(t *testing.T) {
	c := cache.GetLocalCache()
	addTestUserNoID(t, c, "alice")
	addTestUserNoID(t, c, "bob")

	channel := &model.Channel{Id: "chan2", Name: "chan2"}
	responses := make(chan comms.Response, 10)

	originalRoll := rollDiceFn
	defer func() { rollDiceFn = originalRoll }()
	rollDiceFn = (&testRoller{rolls: [][]int{
		{1, 3, 2, 6, 1, 2}, // alice
	}}).roll

	play := func(sender, body string) {
		t.Helper()
		err := BotGame{}.Farkle(BotGame{
			body:            body,
			sender:          sender,
			ReplyChannel:    channel,
			ResponseChannel: responses,
			Cache:           c,
		})
		if err != nil {
			t.Fatalf("call from %s failed: %v", sender, err)
		}
	}

	play("alice", "")
	play("bob", "")
	play("alice", "start")
	play("alice", "roll")
	play("alice", "keep 2 dice")
	play("alice", "bank")

	game, ok, err := loadFarkle(channel.Id, c)
	if err != nil || !ok {
		t.Fatalf("expected farkle game loaded: %v", err)
	}

	aliceKey := playerKey(users.User{Name: "alice"})
	if game.Scores[aliceKey] <= 0 {
		t.Fatalf("expected alice to have points recorded, got %d", game.Scores[aliceKey])
	}
}

func TestFinalRoundSummaryUsesPlayerKey(t *testing.T) {
	c := cache.GetLocalCache()
	addTestUserNoID(t, c, "alice")
	addTestUserNoID(t, c, "bob")

	channel := &model.Channel{Id: "chan3", Name: "chan3"}
	responses := make(chan comms.Response, 20)

	originalRoll := rollDiceFn
	defer func() { rollDiceFn = originalRoll }()
	rollDiceFn = (&testRoller{rolls: [][]int{
		{1, 1, 1, 1, 1, 1}, // alice hits target
		{2, 3, 4, 6, 2, 3}, // bob farkles in final round
	}}).roll

	play := func(sender, body string) {
		t.Helper()
		err := BotGame{}.Farkle(BotGame{
			body:            body,
			sender:          sender,
			ReplyChannel:    channel,
			ResponseChannel: responses,
			Cache:           c,
		})
		if err != nil {
			t.Fatalf("call from %s failed: %v", sender, err)
		}
	}

	play("alice", "")
	play("bob", "")
	play("alice", "start")

	// Lower goal for test brevity.
	game, ok, err := loadFarkle(channel.Id, c)
	if err != nil || !ok {
		t.Fatalf("expected game: %v", err)
	}
	game.TargetScore = 500
	saveFarkle(game, c)

	play("alice", "roll")
	play("alice", "keep 6 dice")
	play("alice", "bank") // triggers final round
	play("bob", "roll")   // farkles, should end game

	close(responses)
	found := false
	for r := range responses {
		if strings.Contains(r.Message, "Game over! Winner: alice") {
			found = true
			if !strings.Contains(r.Message, "alice:") {
				t.Fatalf("expected final scores to include alice entry, got: %s", r.Message)
			}
		}
	}
	if !found {
		t.Fatalf("expected final round summary message")
	}
}

type testRoller struct {
	rolls [][]int
	idx   int
}

func (tr *testRoller) roll(n int) []int {
	if tr.idx >= len(tr.rolls) {
		return rollDice(n)
	}
	roll := tr.rolls[tr.idx]
	tr.idx++
	return roll[:n]
}

func addTestUser(t *testing.T, c cache.Cache, name string) {
	t.Helper()
	u := users.User{Id: name, Name: name}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}
	c.Put(users.KeyPrefix+name, data)
}

func addTestUserNoID(t *testing.T, c cache.Cache, name string) {
	t.Helper()
	u := users.User{Name: name}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}
	c.Put(users.KeyPrefix+name, data)
}
