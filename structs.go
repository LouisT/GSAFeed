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
	OnStart  bool   `json:"onstart,omitempty"`
	ID       string `json:"id"`
	File     string `json:"file"`
	Position string `json:"position,omitempty"`
	Channel  string `json:"channel,omitempty"`
	Killfeed bool   `json:"killfeed,omitempty"`
}

// Geneshift is used to store server metadata
type Geneshift struct {
	Version  string
	Bots     []string
	Killfeed bool
}
