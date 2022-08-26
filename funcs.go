package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hjson/hjson-go"
)

// load a config based on the configFile variable
// use hjson because json does not allow comments
func loadConfig() (Config, error) {
	conf, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	var dat map[string]interface{}
	hjson.Unmarshal(conf, &dat)
	hjsn, err := json.Marshal(dat)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(hjsn, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

// Very basic access levels
func hasAccess(id string, level int64) bool {
	for _, access := range config.Discord.Access {
		if strings.EqualFold(access.ID, id) {
			return level >= access.Level
		}
	}

	return false
}

// TODO: Improve command parser (implement arguments)
func isCommand(input string, cmd string) bool {
	return strings.EqualFold(input, fmt.Sprintf("%s%s", config.Discord.Prefix, cmd))
}
