package main

type Config struct {
	Discord struct {
		Token  string `json:"token"`
		Prefix string `json:"prefix"`
		Access []struct {
			ID    string `json:"id"`
			Level int64  `json:"level"`
		} `json:"access"`
	} `json:"discord"`
	Logs []struct {
		Automatic bool   `json:"automatic,omitempty"`
		ID        string `json:"id"`
		File      string `json:"file"`
		Channel   string `json:"channel,omitempty"`
	} `json:"logs"`
}
