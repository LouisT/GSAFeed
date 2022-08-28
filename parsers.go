package main

import (
	"fmt"
	"regexp"
)

var (
	// Game feed parsers
	// XXX: Implement other game modes.
	//      Make emojis configurable?
	Parsers = map[*regexp.Regexp]func(string, *regexp.Regexp, *Geneshift) (string, bool){
		// Generic parsers
		regexp.MustCompile(`\(\d+\): (.+) joined with steamID: (\d+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(s)
			server.Players[matches[1]] = &Player{ // Always overwrite player stats in case leave is not detected
				Name:    matches[1],
				SteamID: matches[2],
			}
			return fmt.Sprintf(":arrow_right: **%s** has joined the server!", r.FindStringSubmatch(s)[1]), true
		},
		regexp.MustCompile(`(?i)\(\d+\): (HostNewRound|Restarting Match|Queuing Restart Due to New Player)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			server.Bots = append([]string{}, DefaultBots...)
			return "", false
		},
		regexp.MustCompile(`\(\d+\): Saving: (.+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(s)
			delete(server.Players, matches[1])
			return fmt.Sprintf(":arrow_left: **%s** has left the server!", matches[1]), true
		},
		regexp.MustCompile(`\(\d+\): (.+) killed (.+) with (.+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			if !server.Killfeed {
				return "", false
			}
			matches := r.FindStringSubmatch(s)
			isbot1 := ContainsI(server.Bots, matches[1])
			isbot2 := ContainsI(server.Bots, matches[2])
			if isbot1 && isbot2 {
				return "", false
			}
			switch {
			case isbot1:
				if player, ok := server.Players[matches[2]]; ok {
					player.Deaths += 1
					player.KD = divide(player.Kills, player.Deaths)
				}
				matches[1] = fmt.Sprintf("[B] %s", matches[1])
			case isbot2:
				if player, ok := server.Players[matches[2]]; ok {
					player.Kills += 1
					player.KD = divide(player.Kills, player.Deaths)
				}
				matches[2] = fmt.Sprintf("[B] %s", matches[2])
			}
			return fmt.Sprintf(":skull_crossbones: **%s** has killed **%s** with a **%s**", matches[1], matches[2], matches[3]), true
		},

		// BR parsers
		regexp.MustCompile(`\(\d+\): SERVER: (.+) wins round (\d+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(s)
			if ContainsI(server.Bots, matches[1]) {
				return fmt.Sprintf(":person_facepalming: This is a sad day... everyone lost to **[B] %s**, a bot, on round **%s**!", matches[1], matches[2]), true
			}
			return fmt.Sprintf(":trophy: ***%s*** has won round **%s**!", matches[1], matches[2]), true
		},
		regexp.MustCompile(`\(\d+\): SERVER: (.+) gets the winner winner duck dinner`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(s)
			if ContainsI(server.Bots, matches[1]) {
				return fmt.Sprintf(":person_facepalming: Boooo! Everyone lost to **[B] %s**, a bot, on the ***FINAL*** round!", matches[1]), true
			}
			return fmt.Sprintf(":poultry_leg: Congratulations to ***%s***! They have won the ***FINAL*** round!", matches[1]), true
		},
		regexp.MustCompile(`\(\d+\): RestartBattleRound: : (\d+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			return fmt.Sprintf(":exclamation: Everyone get ready, round **%s** is starting!", r.FindStringSubmatch(s)[1]), true
		},
	}

	// List of parsers used to track/update server settings
	MetaParsers = map[string]*regexp.Regexp{
		"StartLine":    regexp.MustCompile(`\(\d+\): ========= Start Loading Geneshift (.[^\s]+)`),          // Get Geneshift server version
		"Reset":        regexp.MustCompile(`\(\d+\): (Restarting Match|Queuing Restart Due to New Player)`), // Clear bot list?
		"AddPlayer":    regexp.MustCompile(`\(\d+\): (.+) joined with steamID: (\d+)`),
		"RemovePlayer": regexp.MustCompile(`\(\d+\): Saving: (.+)`),
		"AddBot":       regexp.MustCompile(`\(\d+\): Adding Bot: (.+) with target`),
		"Kills":        regexp.MustCompile(`\(\d+\): (.+) killed (.+) with (.+)`),
	}
)
