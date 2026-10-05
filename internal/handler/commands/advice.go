package commands

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/jokes"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

type adviceProvider interface {
	Advice(context.Context, string) (string, error)
}

func (h BotCommandHelp) Advice(request BotCommand) (response HelpResponse) {
	response.Help = `Ask Bender for advice on a topic or question. Usage: '!advice <question>' (e.g. '!advice Could you help me understand binary numbering?')`
	response.Description = `Get advice from Bender's sarcastic robot perspective.`
	return response
}

func (bc BotCommand) Advice(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "post",
	}
	question := strings.TrimSpace(event.body)
	if question == "" {
		response.Type = "dm"
		response.Message = "Give me a question, meatbag. Try `!advice Could you help me understand binary numbering?`"
		event.ResponseChannel <- response
		return nil
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		response.Message = "My advice circuits are offline. Set OPENAI_API_KEY and restart me."
		event.ResponseChannel <- response
		return nil
	}
	provider := jokes.OpenAI{
		Client: &http.Client{Timeout: 30 * time.Second},
		APIKey: apiKey,
		Model:  os.Getenv("OPENAI_MODEL"),
	}
	return respondWithAdvice(event, response, provider, question)
}

func respondWithAdvice(event BotCommand, response comms.Response, provider adviceProvider, question string) error {
	answer, err := provider.Advice(context.Background(), question)
	if err != nil {
		log.Printf("Advice command: %v", err)
		response.Message = "My advice circuits are smoking. Try again in a bit."
	} else {
		response.Message = answer
	}
	event.ResponseChannel <- response
	return nil
}
