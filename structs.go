package main

// Config is the overall config file
type Config struct {
	Discord struct {
		Avatar struct {
			URL    string `json:"url"`
			Update bool   `json:"update"`
		} `json:"avatar"`
		Token    string   `json:"token"`
		Prefixes string   `json:"prefixes"`
		Channels []string `json:"channels"`
		Access   []struct {
			ID    string `json:"id"`
			Level int64  `json:"level"`
		} `json:"access"`
	} `json:"discord"`
	Logs []Logs `json:"logs"`
}

// Logs are the log file settings
type Logs struct {
	Preload  bool   `json:"preload"`
	OnStart  bool   `json:"onstart,omitempty"`
	ID       string `json:"id"`
	File     string `json:"file"`
	Position string `json:"position,omitempty"`
	Channel  string `json:"channel,omitempty"`
	Killfeed bool   `json:"killfeed,omitempty"`
}

// Geneshift is used to store server metadata
type Geneshift struct {
	CanEmit   bool // If true, send channel messages
	Version   string
	Players   map[string]*Player
	Bots      []string
	RoundWins map[int]string
	Killfeed  bool
	Finished  bool
}

// Track players
type Player struct {
	Name    string
	SteamID string // Not currently tracked?
	Kills   int
	Deaths  int
	KD      float64
}

// Reset a player stats
func (p *Player) reset() *Player {
	p.Kills = 0
	p.Deaths = 0
	p.KD = 0

	return p
}
