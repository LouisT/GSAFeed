package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hjson/hjson-go"
	"github.com/nxadm/tail"
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

// HasAccess provides basic access levels for specific discord user IDs
func HasAccess(id string, level int64) bool {
	for _, access := range config.Discord.Access {
		if strings.EqualFold(access.ID, id) {
			return level >= access.Level
		}
	}

	return false
}

// IsCommand checks input string for prefix + command
func IsCommand(input string, cmd string) bool {
	return IsCommandPrefix(input, cmd, config.Discord.Prefixes)
}

// IsCommandPrefix checks if a command has specific prefixes
func IsCommandPrefix(input string, cmd string, prefix string) bool {
	return regexp.MustCompile(fmt.Sprintf("^(?i)[%s]%s", prefix, cmd)).MatchString(input)
}

// GetCommand returns ["prefix", "command", "arguments"]
func GetCommand(input string) (string, string, string) {
	split := regexp.MustCompile(`\s`).Split(strings.TrimSpace(input), 2)

	if len(split) == 2 {
		return split[0][:1], split[0][1:], split[1]
	}

	return split[0][:1], split[0][1:], ""
}

// LogParser manages the log parsing for Geneshift logs
func LogParser(session *discordgo.Session, settings Logs) {
	if _, err := session.ChannelMessageSend(settings.Channel, fmt.Sprintf("***=== Starting game stream! (ID: %s) ===***", settings.ID)); err != nil {
		log.Printf("[%s/%s] Message error: %+v", settings.ID, settings.Channel, err)
	}
	Whence := 2
	switch settings.Position {
	case "start":
		Whence = io.SeekStart
	default:
		Whence = io.SeekEnd
	}
	if tailer, err := tail.TailFile(settings.File, tail.Config{
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: Whence,
		},
		Follow: true,
		ReOpen: true, // If the Geneshift server restarts, keep reading the file after trunc.
	}); err == nil {
		Tails[settings.ID] = tailer // Append for Cleanup()
		go func() {
			for line := range tailer.Lines {
				for rgx, compiler := range Parsers {
					if rgx.MatchString(line.Text) {
						if _, err := session.ChannelMessageSend(settings.Channel, fmt.Sprintf("[%s] %s", settings.ID, compiler(line.Text, rgx))); err != nil {
							log.Printf("[%s] Message error: %+v", settings.Channel, err)
						}
					}
				}
			}
		}()
	} else {
		log.Println(err)
	}
}
