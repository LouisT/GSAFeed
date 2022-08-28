package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/hjson/hjson-go"
	"github.com/nxadm/tail"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

// MessageParser manages the log parsing for Geneshift logs
func MessageParser(session *discordgo.Session, settings Logs) {
	if _, ok := Onces[settings.ID]; !ok {
		Onces[settings.ID] = &sync.Once{}
	}

	Onces[settings.ID].Do(func() {
		var GSSettings *Geneshift
		var err error
		if GSSettings, err = Preload(settings); err != nil {
			log.Println(err)
		}
		Servers[settings.ID] = GSSettings
		msg := fmt.Sprintf("***>>> Starting game feed for Geneshift %s (ID: %s)***", Servers[settings.ID].Version, settings.ID)
		if _, err := session.ChannelMessageSend(settings.Channel, normalize(msg)); err != nil {
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
				var last string
				for line := range tailer.Lines {
					if MetaParsers["Reset"].MatchString(line.Text) {
						Servers[settings.ID].Bots = append([]string{}, DefaultBots...)
					} else if MetaParsers["AddBot"].MatchString(line.Text) {
						if match := MetaParsers["AddBot"].FindStringSubmatch(line.Text); len(match) == 2 {
							Servers[settings.ID].Bots = append(Servers[settings.ID].Bots, match[1])
						}
					} else {
						for rgx, compiler := range Parsers {
							if last != line.Text && rgx.MatchString(line.Text) {
								if output, ok := compiler(line.Text, rgx, Servers[settings.ID]); ok && output != last {
									if Servers[settings.ID].CanEmit {
										if _, err := session.ChannelMessageSend(settings.Channel, fmt.Sprintf("[%s] %s", settings.ID, output)); err != nil {
											log.Printf("[%s] Message error: %+v", settings.Channel, err)
										}
									}
									last = output
								}
							}
						}
					}
				}
			}()
		} else {
			log.Println(err)
		}
	})
}

// PreloadLog attempts to read the log file as fast as possible,
// prefilling bots and player data.
func Preload(opts Logs) (*Geneshift, error) {
	settings := &Geneshift{
		Players:  make(map[string]*Player),
		Bots:     append([]string{}, DefaultBots...),
		Killfeed: opts.Killfeed,
	}
	f, err := os.Open(opts.File)
	if err != nil {
		return settings, err
	}
	defer f.Close()
	lines := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines++
		txt := scanner.Text()
		if lines == 1 {
			if !MetaParsers["StartLine"].MatchString(txt) {
				return settings, fmt.Errorf("---!--- invalid start of log file, no settings found")
			} else {
				if match := MetaParsers["StartLine"].FindStringSubmatch(txt); len(match) == 2 {
					settings.Version = match[1]
				}
			}
		} else if MetaParsers["Reset"].MatchString(txt) {
			settings.Bots = append([]string{}, DefaultBots...)
		} else if MetaParsers["Kills"].MatchString(txt) {
			if match := MetaParsers["Kills"].FindStringSubmatch(txt); len(match) == 4 {
				if player, ok := settings.Players[match[1]]; ok {
					player.Kills += 1
					player.KD = divide(player.Kills, player.Deaths)
				}
				if player, ok := settings.Players[match[2]]; ok {
					player.Deaths += 1
					player.KD = divide(player.Kills, player.Deaths)
				}
			}
		} else if MetaParsers["AddPlayer"].MatchString(txt) {
			if match := MetaParsers["AddPlayer"].FindStringSubmatch(txt); len(match) == 3 {
				settings.Players[match[1]] = &Player{ // Always overwrite player stats in case leave is not detected
					Name:    match[1],
					SteamID: match[2],
				}
			}
		} else if MetaParsers["RemovePlayer"].MatchString(txt) {
			if match := MetaParsers["RemovePlayer"].FindStringSubmatch(txt); len(match) == 2 {
				delete(settings.Players, match[1])
			}
		} else if MetaParsers["AddBot"].MatchString(txt) {
			if match := MetaParsers["AddBot"].FindStringSubmatch(txt); len(match) == 2 {
				settings.Bots = append(settings.Bots, match[1])
			}
		} else if !opts.Preload && strings.Contains(txt, "Finish Loading Sequence") {
			return settings, nil
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
	p := message.NewPrinter(language.English)
	logger.Printf(">>> Preloaded %s lines! <<<", p.Sprintf("%d", lines))
	settings.CanEmit = true

	return settings, nil
}

// ContainsI checks if a string exists in a string slice
func ContainsI(slice []string, key string) bool {
	for _, value := range slice {
		if strings.EqualFold(value, key) {
			return true
		}
	}
	return false
}

// normalize removes any extra spaces or other characters (in the future)
func normalize(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

// Calculate KD with zero error catch
func divide(numerator int, denominator int) (result float64) {
	defer func() {
		if r := recover(); r != nil {
			result = float64(numerator)
		}
	}()
	return float64(numerator / denominator)
}
