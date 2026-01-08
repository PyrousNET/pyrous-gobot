package settings

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type (
	Settings struct {
		mu          sync.RWMutex
		settingsUrl string
		settings    CommandSettings
	}

	CommandSettings struct {
		CommandTrigger string              `json:"command_start"`
		GameTrigger    string              `json:"game_start"`
		Insults        []string            `json:"insults"`
		Quotes         []string            `json:"quotes"`
		Praises        []string            `json:"praises"`
		Reactions      map[string]Reaction `json:"reactions"`
	}

	Reaction struct {
		Url         string `json:"url"`
		Description string `json:"description"`
	}
)

func NewSettings(settingsUrl string) (*Settings, error) {
	sc := &Settings{
		settingsUrl: settingsUrl,
	}

	err := sc.LoadSettings()

	return sc, err
}

func SetupMockSettings(mu sync.RWMutex, settings CommandSettings) *Settings {
	c := &Settings{}
	c.mu = mu
	c.settings = settings
	return c
}

func (c *Settings) LoadSettings() error {
	var s CommandSettings
	hc := &http.Client{Timeout: 10 * time.Second}

	r, err := hc.Get(c.settingsUrl)
	if err != nil {
		c.mu.Lock()
		c.settings = s
		c.loadLocalDefaults()
		c.mu.Unlock()
		return nil
	}
	defer r.Body.Close()

	if r.StatusCode < 200 || r.StatusCode >= 300 {
		c.mu.Lock()
		c.settings = s
		c.loadLocalDefaults()
		c.mu.Unlock()
		return nil
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.mu.Lock()
		c.settings = s
		c.loadLocalDefaults()
		c.mu.Unlock()
		return nil
	}

	err = json.Unmarshal(b, &s)
	if err != nil {
		c.mu.Lock()
		c.settings = s
		c.loadLocalDefaults()
		c.mu.Unlock()
		return nil
	}

	c.mu.Lock()
	c.settings = s

	// this is temporary until we get these pulling from the server
	c.loadLocalDefaults()
	c.mu.Unlock()

	return nil
}

func (c *Settings) GetCommandTrigger() string {
	c.mu.RLock()
	commandTrigger := c.settings.CommandTrigger
	c.mu.RUnlock()
	return normalizeTrigger(commandTrigger, "!")
}

func (c *Settings) GetGameTrigger() string {
	c.mu.RLock()
	gameTrigger := c.settings.GameTrigger
	c.mu.RUnlock()
	return normalizeTrigger(gameTrigger, "$")
}

func normalizeTrigger(trigger string, defaultValue string) string {
	value := strings.TrimSpace(trigger)
	if strings.HasPrefix(value, "^") {
		value = strings.TrimPrefix(value, "^")
	}
	if strings.HasPrefix(value, "\\") && len(value) > 1 {
		value = value[1:]
	}
	if strings.HasSuffix(value, "$") && len(value) > 1 {
		value = strings.TrimSuffix(value, "$")
	}
	if strings.HasPrefix(value, "\\") && len(value) > 1 {
		value = value[1:]
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Settings) GetInsults() []string {
	c.mu.RLock()
	insults := c.settings.Insults
	c.mu.RUnlock()
	if len(insults) == 0 {
		c.mu.Lock()
		if len(c.settings.Insults) == 0 {
			c.settings.Insults = loadLocalStringList("insults.json")
		}
		insults = c.settings.Insults
		c.mu.Unlock()
	}
	return insults
}

func (c *Settings) GetQuotes() []string {
	c.mu.RLock()
	quotes := c.settings.Quotes
	c.mu.RUnlock()
	if len(quotes) == 0 {
		c.mu.Lock()
		if len(c.settings.Quotes) == 0 {
			c.settings.Quotes = loadLocalStringList("quotes.json")
		}
		quotes = c.settings.Quotes
		c.mu.Unlock()
	}
	return quotes
}

func (c *Settings) GetPraises() []string {
	c.mu.RLock()
	praises := c.settings.Praises
	c.mu.RUnlock()
	if len(praises) == 0 {
		c.mu.Lock()
		if len(c.settings.Praises) == 0 {
			c.settings.Praises = loadLocalStringList("praises.json")
		}
		praises = c.settings.Praises
		c.mu.Unlock()
	}
	return praises
}

func (c *Settings) GetReactions() map[string]Reaction {
	c.mu.RLock()
	reactions := c.settings.Reactions
	c.mu.RUnlock()
	if len(reactions) == 0 {
		c.mu.Lock()
		if len(c.settings.Reactions) == 0 {
			c.loadLocalReactions()
		}
		reactions = c.settings.Reactions
		c.mu.Unlock()
	}
	return reactions
}

func (c *Settings) loadLocalReactions() {
	b, err := os.ReadFile("./reactions.json")
	if err != nil {
		return
	}

	var reactions map[string]Reaction
	if err := json.Unmarshal(b, &reactions); err != nil {
		return
	}

	c.settings.Reactions = reactions
}

func loadLocalStringList(filename string) []string {
	b, err := os.ReadFile("./" + filename)
	if err != nil {
		return nil
	}

	var values []string
	if err := json.Unmarshal(b, &values); err != nil {
		return nil
	}

	return values
}

func (c *Settings) loadLocalDefaults() {
	if len(c.settings.Reactions) == 0 {
		c.loadLocalReactions()
	}
	if len(c.settings.Insults) == 0 {
		c.settings.Insults = loadLocalStringList("insults.json")
	}
	if len(c.settings.Quotes) == 0 {
		c.settings.Quotes = loadLocalStringList("quotes.json")
	}
	if len(c.settings.Praises) == 0 {
		c.settings.Praises = loadLocalStringList("praises.json")
	}
}
