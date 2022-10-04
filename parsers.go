package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	// Game feed parsers
	// XXX: Implement other game modes.
	//      Make emojis configurable?
	Parsers = map[*regexp.Regexp]func(*discordgo.Session, Logs, string, *regexp.Regexp, *Geneshift) (string, bool){
		// Generic parsers
		regexp.MustCompile(`\(\d+\): (.+) joined with steamID: (\d+)`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(str)
			prefix := "re"
			if _, exists := server.Players[matches[1]]; !exists {
				server.Players[matches[1]] = &Player{Name: matches[1]}
				prefix = ""
			}
			return fmt.Sprintf(":arrow_right: **%s** has %sjoined the server!", matches[1], prefix), true
		},
		regexp.MustCompile(`(?i)\(\d+\): Sending Round Over`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			if server.Finished {
				server.reset(false)
			}
			return "", false
		},
		regexp.MustCompile(`(?i)\(\d+\): (HostNewRound|Restarting Match|Queuing Restart Due to New Player)`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			server.reset(false)
			return "", false
		},
		regexp.MustCompile(`\(\d+\): Saving: (.+)`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(str)
			// XXX: Keep tracking for top players in a match?
			// delete(server.Players, matches[1])
			return fmt.Sprintf(":arrow_left: **%s** has left the server!", matches[1]), true
		},
		regexp.MustCompile(`\(\d+\): (.+) killed (.+) with (.+)`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(str)
			if ContainsI(server.Bots, matches[1]) && ContainsI(server.Bots, matches[2]) {
				return "", false
			}
			if player, ok := server.Players[matches[1]]; ok {
				player.Kills += 1
				player.KD = KD(player)
			} else {
				matches[1] = fmt.Sprintf("[B] %s", matches[1])
			}
			if player, ok := server.Players[matches[2]]; ok {
				player.Deaths += 1
				player.KD = KD(player)
			} else {
				matches[2] = fmt.Sprintf("[B] %s", matches[2])
			}
			if !server.Killfeed {
				return "", false
			}
			return fmt.Sprintf(":skull_crossbones: **%s** has killed **%s** with a **%s**", matches[1], matches[2], matches[3]), true
		},

		// BR parsers
		regexp.MustCompile(`\(\d+\): SERVER: (.+) wins round (\d+)`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(str)
			name, isbot := func(n string) (string, bool) { // Normalize player/bot name
				for key := range server.Players {
					if strings.EqualFold(n, key) {
						return key, false
					}
				}
				for _, key := range server.Bots {
					if strings.EqualFold(n, key) {
						return fmt.Sprintf("[B] %s", key), true
					}
				}
				return n, false // Unknown user?
			}(matches[1])
			round, _ := strconv.Atoi(matches[2])
			server.RoundWins[round] = name
			if isbot {
				return fmt.Sprintf(":robot: **%s**, a bot, has won round **%s**!", name, matches[2]), true
			}
			return fmt.Sprintf(":adult: ***%s*** has won round **%s**!", name, matches[2]), true
		},
		regexp.MustCompile(`\(\d+\): SERVER: (.+) gets the winner winner`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			if server.Finished {
				return "", false
			}
			server.Finished = true
			matches := r.FindStringSubmatch(str)
			fields := []*discordgo.MessageEmbedField{}
			keys := make([]string, 0, len(server.Players))
			for k := range server.Players {
				keys = append(keys, k)
			}
			sort.SliceStable(keys, func(i, j int) bool {
				return server.Players[keys[i]].KD > server.Players[keys[j]].KD
			})
			if len(keys) >= 10 {
				keys = keys[:10]
			}
			kds := []string{}
			for _, k := range keys {
				player := server.Players[k]
				kds = append(kds, fmt.Sprintf("**%s** %d/%d (%0.3f)", player.Name, player.Kills, player.Deaths, player.KD))
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Top ***%d*** player%s:", len(kds), map[bool]string{true: "", false: "s"}[len(kds) == 1]),
				Value:  strings.Join(kds, ":black_small_square: "),
				Inline: false,
			})
			rounds := make([]int, 0, len(server.RoundWins))
			for round := range server.RoundWins {
				rounds = append(rounds, round)
			}
			sort.Ints(rounds)
			for _, round := range rounds {
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Round ***%d*** winner:", round),
					Value:  server.RoundWins[round],
					Inline: true,
				})
			}
			title := fmt.Sprintf(":trophy: Congratulations, ***%s***! They won the ***FINAL*** round!", matches[1])
			if ContainsI(server.Bots, matches[1]) {
				title = fmt.Sprintf(":person_facepalming: A bot, **[B] %s**, has won the ***FINAL*** round!", matches[1])
			}
			if _, err := session.ChannelMessageSendEmbed(settings.Channel, &discordgo.MessageEmbed{
				Author:    &discordgo.MessageEmbedAuthor{},
				Color:     0x00ff00, // Green
				Fields:    fields,
				Timestamp: time.Now().Format(time.RFC3339),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://i.imgur.com/575ebif.gif",
				},
				Title: normalize(title),
				Footer: &discordgo.MessageEmbedFooter{
					Text: normalize(fmt.Sprintf("%s - Geneshift %v", settings.ID, server.Version)),
				},
			}); err != nil {
				log.Println(err)
			}
			for _, player := range server.Players {
				player.Reset().selfdestruct(server.Players)
			}
			return "", false
		},
		regexp.MustCompile(`\(\d+\): RestartBattleRound: : (\d+)`): func(session *discordgo.Session, settings Logs, str string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			return fmt.Sprintf(":exclamation: Round **%s** is starting!", r.FindStringSubmatch(str)[1]), true
		},
	}

	// List of parsers used to track/update server settings
	MetaParsers = map[string]*regexp.Regexp{
		"StartLine":    regexp.MustCompile(`\(\d+\): ========= Start Loading Geneshift (.[^\s]+)`),                           // Get Geneshift server version
		"Reset":        regexp.MustCompile(`(?i)\(\d+\): (HostNewRound|Restarting Match|Queuing Restart Due to New Player)`), // Clear bot list?
		"AddPlayer":    regexp.MustCompile(`\(\d+\): (.+) joined with steamID: (\d+)`),
		"RemovePlayer": regexp.MustCompile(`\(\d+\): Saving: (.+)`),
		"AddBot":       regexp.MustCompile(`\(\d+\): Adding Bot: (.+) with target`),
		"Kills":        regexp.MustCompile(`\(\d+\): (.+) killed (.+) with (.+)`),
	}
)
