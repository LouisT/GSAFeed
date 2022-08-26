package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	configFile = ""
	config     Config
	Once       = &sync.Once{}
)

func init() {
	// Set logging to stdout
	log.SetOutput(os.Stdout)
}

func main() {
	flag.StringVar(&configFile, "config", "./config.hjson", "path to config file")
	flag.Parse()

	var err error
	config, err = loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	dg, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	if err = dg.Open(); err != nil {
		log.Fatal("error opening connection,", err)
		return
	}
	defer dg.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if isCommand(m.Content, "myinfo") && hasAccess(m.Author.ID, 1) {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("ChannelID: %s, AuthorID: %s", m.ChannelID, m.Author.ID))
	}
}
