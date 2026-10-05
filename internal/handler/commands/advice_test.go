package commands

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

type testAdviceProvider struct {
	answer   string
	err      error
	question string
	calls    int
}

func (provider *testAdviceProvider) Advice(_ context.Context, question string) (string, error) {
	provider.calls++
	provider.question = question
	return provider.answer, provider.err
}

func TestRespondWithAdvicePostsAnswerToRequestChannel(t *testing.T) {
	responses := make(chan comms.Response, 1)
	event := BotCommand{ResponseChannel: responses}
	provider := &testAdviceProvider{answer: "Binary uses powers of two, meatbag."}
	response := comms.Response{ReplyChannelId: "channel", Type: "post", UserId: "user"}

	if err := respondWithAdvice(event, response, provider, "Could you explain binary numbering?"); err != nil {
		t.Fatalf("respondWithAdvice() error = %v", err)
	}
	got := <-responses
	if got.Message != provider.answer || got.Type != "post" || got.ReplyChannelId != "channel" || got.UserId != "user" {
		t.Fatalf("response = %#v, want channel post containing advice", got)
	}
	if provider.calls != 1 || provider.question != "Could you explain binary numbering?" {
		t.Fatalf("provider calls/question = %d/%q", provider.calls, provider.question)
	}
}

func TestRespondWithAdvicePostsFriendlyFailureWhenProviderFails(t *testing.T) {
	responses := make(chan comms.Response, 1)
	event := BotCommand{ResponseChannel: responses}
	provider := &testAdviceProvider{err: errors.New("service unavailable")}

	if err := respondWithAdvice(event, comms.Response{Type: "post"}, provider, "Question"); err != nil {
		t.Fatalf("respondWithAdvice() error = %v", err)
	}
	got := <-responses
	if got.Type != "post" || !strings.Contains(got.Message, "advice circuits") {
		t.Fatalf("failure response = %#v, want a friendly channel reply", got)
	}
}

func TestAdviceHelpDescribesUsageAndCommandIsRegistered(t *testing.T) {
	doc := (BotCommandHelp{}).Advice(BotCommand{})
	if !strings.Contains(doc.Help, "!advice <question>") || !strings.Contains(strings.ToLower(doc.Description), "advice") {
		t.Fatalf("advice help = %#v, want usage and description", doc)
	}

	sttngs := settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{})
	commands := NewCommands(sttngs, nil, &cache.MockCache{}, nil)
	if _, err := commands.getMethod("Advice"); err != nil {
		t.Fatalf("Advice command was not registered: %v", err)
	}
}

func TestAdviceReturnsQuestionHelpInDMWhenEmpty(t *testing.T) {
	responses := make(chan comms.Response, 1)
	event := BotCommand{
		body:            "  ",
		sender:          "@test-user",
		ReplyChannel:    &model.Channel{Id: "channel"},
		ResponseChannel: responses,
		cache:           &cache.MockCache{},
	}
	if err := (BotCommand{}).Advice(event); err != nil {
		t.Fatalf("Advice() error = %v", err)
	}
	got := <-responses
	if got.Type != "dm" || !strings.Contains(got.Message, "!advice") {
		t.Fatalf("empty-question response = %#v, want DM with usage", got)
	}
}

func TestAdviceWithoutAPIKeyPostsConfigurationHint(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	responses := make(chan comms.Response, 1)
	event := BotCommand{
		body:            "Explain binary numbering",
		sender:          "@test-user",
		ReplyChannel:    &model.Channel{Id: "channel"},
		ResponseChannel: responses,
		cache:           &cache.MockCache{},
	}
	if err := (BotCommand{}).Advice(event); err != nil {
		t.Fatalf("Advice() error = %v", err)
	}
	got := <-responses
	if got.Type != "post" || !strings.Contains(got.Message, "OPENAI_API_KEY") {
		t.Fatalf("missing-key response = %#v, want channel post explaining configuration", got)
	}
}
