package main

import (
	"fmt"
	"regexp"
)

var (
	// Log parsers
	// XXX: Look for a better method of parsing lines. (no more regexp?)
	//      Implement other game modes.
	//      Make emojis configurable?
	Parsers = map[*regexp.Regexp]func(string, *regexp.Regexp) string{
		regexp.MustCompile(`: RestartBattleRound: : (\d+)`): func(s string, r *regexp.Regexp) string {
			return fmt.Sprintf(":exclamation: Round **%s** is starting!", r.FindStringSubmatch(s)[1])
		},
		regexp.MustCompile(`: Validation Successful: (.[^\d]+) (?:\d+)`): func(s string, r *regexp.Regexp) string {
			return fmt.Sprintf(":arrow_right: **%s** has joined the server!", r.FindStringSubmatch(s)[1])
		},
		regexp.MustCompile(`: Saving: (.+)`): func(s string, r *regexp.Regexp) string {
			return fmt.Sprintf(":arrow_left: **%s** has left the server!", r.FindStringSubmatch(s)[1])
		},
		regexp.MustCompile(`: (.+) killed (.+) with (.+)`): func(s string, r *regexp.Regexp) string {
			matches := r.FindStringSubmatch(s)
			return fmt.Sprintf(":skull_crossbones: **%s** has killed **%s** with a **%s**", matches[1], matches[2], matches[3])
		},
		regexp.MustCompile(`: SERVER: (.+) wins round (\d+)`): func(s string, r *regexp.Regexp) string {
			matches := r.FindStringSubmatch(s)
			return fmt.Sprintf(":trophy: Congratulations, **%s** has won round **%s**!", matches[1], matches[2])
		},
	}
)
