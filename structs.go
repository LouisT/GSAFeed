package main

import "time"

// Config is the overall config file
type Config struct {
	GSA struct {
		Servers string `json:"servers"`
		Bots    string `json:"bots"`
	} `json:"gsa"`
	Discord struct {
		Avatar struct {
			File   string `json:"file"`
			URL    string `json:"url"`
			Update bool   `json:"update"`
		} `json:"avatar"`
		Token    string   `json:"token"`
		Prefixes []string `json:"prefixes"`
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

// GSA is used to store server metadata
type GSA struct {
	CanEmit   bool // If true, send channel messages
	Version   string
	Players   map[string]*Player
	Bots      []string
	RoundWins map[int]string
	Killfeed  bool
	Finished  bool
}

// NewGSA creates a server instance
func NewGSA() *GSA {
	return &GSA{
		RoundWins: make(map[int]string),
		Players:   make(map[string]*Player),
	}
}

// Reset Server metadata
func (g *GSA) reset(players bool) *GSA {
	g.Finished = false
	g.RoundWins = make(map[int]string)
	if players {
		g.Players = make(map[string]*Player)
	}

	return g
}

// Track players
type Player struct {
	Name    string
	SteamID string // Not currently tracked?
	Kills   int
	Deaths  int
	KD      float64
	timer   *time.Timer
}

// Reset a player stats
func (p *Player) Reset() *Player {
	if p.timer != nil {
		p.timer.Stop()
	}
	p.Kills = 0
	p.Deaths = 0
	p.KD = 0

	return p
}

// selfdestruct removes a player from the list in 5 minutes
// if they don't rejoin the server after a match
func (p *Player) selfdestruct(list map[string]*Player) *Player {
	p.timer = time.AfterFunc(time.Minute*5, func() {
		delete(list, p.Name)
	})

	return p
}
