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
		regexp.MustCompile(`: Validation Successful: (.[^\d]+) (?:\d+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			return fmt.Sprintf(":arrow_right: **%s** has joined the server!", r.FindStringSubmatch(s)[1]), true
		},
		regexp.MustCompile(`: Saving: (.+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			return fmt.Sprintf(":arrow_left: **%s** has left the server!", r.FindStringSubmatch(s)[1]), true
		},
		regexp.MustCompile(`: (.+) killed (.+) with (.+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
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
				matches[1] = fmt.Sprintf("[B] %s", matches[1])
			case isbot2:
				matches[2] = fmt.Sprintf("[B] %s", matches[2])
			}
			return fmt.Sprintf(":skull_crossbones: **%s** has killed **%s** with a **%s**", matches[1], matches[2], matches[3]), true
		},

		// BR parsers
		regexp.MustCompile(`: SERVER: (.+) wins round (\d+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(s)
			if ContainsI(server.Bots, matches[1]) {
				return fmt.Sprintf(":person_facepalming: This is a sad day... everyone lost to **[B] %s**, a bot, on round **%s**!", matches[1], matches[2]), true
			}
			return fmt.Sprintf(":trophy: ***%s*** has won round **%s**!", matches[1], matches[2]), true
		},
		regexp.MustCompile(`: SERVER: (.+) gets the winner winner duck dinner`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			matches := r.FindStringSubmatch(s)
			if ContainsI(server.Bots, matches[1]) {
				return fmt.Sprintf(":person_facepalming: Boooo! Everyone lost to **[B] %s**, a bot, on the ***FINAL*** round!", matches[1]), true
			}
			return fmt.Sprintf(":poultry_leg: Congratulations to ***%s***! They have won the ***FINAL*** round!", matches[1]), true
		},
		regexp.MustCompile(`: RestartBattleRound: : (\d+)`): func(s string, r *regexp.Regexp, server *Geneshift) (string, bool) {
			return fmt.Sprintf(":exclamation: Everyone get ready, round **%s** is starting!", r.FindStringSubmatch(s)[1]), true
		},
	}

	// List of parsers used to track/update server settings
	MetaParsers = map[string]*regexp.Regexp{
		"StartLine": regexp.MustCompile(`: ========= Start Loading Geneshift (.[^\s]+)`), // Get Geneshift server version
		"Reset":     regexp.MustCompile(`: Restarting Match`),                            // Clear bot list?
		"AddBot":    regexp.MustCompile(`: Adding Bot: (.+) with target`),
	}
)
