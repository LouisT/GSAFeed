package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/nxadm/tail"
)

var (
	configFile = ""
	config     Config
	Onces      map[string]*sync.Once = make(map[string]*sync.Once)
	Allowed    map[string]bool       = make(map[string]bool)
	Tails      map[string]*tail.Tail = make(map[string]*tail.Tail)
	logger                           = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// List of bots for specific server IDs
	DefaultBots = []string{ "a civilian" }
	Bots map[string][]string = make(map[string][]string)
)

func main() {
	defer func() {
		for id, tail := range Tails {
			logger.Printf("Clearning up %s\n", id)
			tail.Cleanup() // Cleanup tails from inotify
		}
	}()

	flag.StringVar(&configFile, "config", "./config.hjson", "path to config file")
	flag.Parse()

	var err error
	if config, err = loadConfig(); err != nil {
		log.Fatal(err)
	}

	dg, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
		return
	}

	// Store "allowed" chnannels for command permissions
	for _, channel := range config.Discord.Channels {
		Allowed[channel] = true
	}

	// Set the bot avatar
	if config.Discord.Avatar.Update && len(config.Discord.Avatar.URL) > 1 {
		logger.Println("Updating avatar...")
		go func() {
			resp, err := http.Get(config.Discord.Avatar.URL)
			if err != nil {
				log.Println("Error retrieving the file, ", err)
				return
			}
			defer resp.Body.Close()
			img, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error reading the response, ", err)
				return
			}
			if _, err = dg.UserUpdate("", fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64.StdEncoding.EncodeToString(img))); err != nil {
				log.Println(err)
			}
		}()
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if _, ok := Allowed[m.ChannelID]; !ok || m.Author.ID == s.State.User.ID {
			return
		}

		// Returns basic information about a Discord user
		if IsCommand(m.Content, "myinfo") {
			_, cmd, args := GetCommand(m.Content)
			if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("[%s] %s / ChannelID: %s, AuthorID: %s", cmd, args, m.ChannelID, m.Author.ID)); err != nil {
				log.Printf("[%s] Message error: %+v", m.ChannelID, err)
			}
		}

		// TODO: Add !start, !stop commands for tailing.
	})
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	if err = dg.Open(); err != nil {
		log.Fatal("error opening connection,", err)
		return
	}
	defer dg.Close()

	for _, settings := range config.Logs {
		if GSSettings, err := GeneshiftSettings(settings.File); err == nil {
			Bots[settings.ID] = GSSettings.Bots
		} else {
			log.Println(err)
		}
		if settings.OnStart {
			go LogParser(dg, settings)
		}
	}

	logger.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
