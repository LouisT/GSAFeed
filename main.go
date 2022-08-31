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
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/nxadm/tail"
)

var (
	// Project is set at compile time
	Project = "GSFeed"
	// Version is set at compile time
	Version = "0.0.0-beta.0"
	// Revision is set at compile time, it is the git SHA-1 revision
	Revision = "0000000"

	configFile = ""
	config     Config
	Onces      map[string]*sync.Once = make(map[string]*sync.Once)
	Allowed    map[string]bool       = make(map[string]bool)
	Tails      map[string]*tail.Tail = make(map[string]*tail.Tail)
	logger                           = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	IDs                              = []string{}

	// List of bots for specific server IDs
	DefaultBots = []string{"a civilian"}

	// Geneshift server metadata
	Servers map[string]*Geneshift = make(map[string]*Geneshift)
)

func cleanup() {
	for id, tail := range Tails {
		logger.Printf("Cleaning %s\n", id)
		tail.Cleanup() // Cleanup tails from inotify
	}
}

func main() {
	logger.Printf("%s v%s+%s - %s\n", Project, Version, Revision, runtime.Version())

	defer cleanup()
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
		if IsCommand(m.Content, "myinfo") {
			_, cmd, args := GetCommand(m.Content)
			if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("[%s] %s / ChannelID: %s, AuthorID: %s", cmd, args, m.ChannelID, m.Author.ID)); err != nil {
				log.Printf("[%s] Message error: %+v", m.ChannelID, err)
			}
		} else if HasAccess(m.Author.ID, 1) {
			if IsCommand(m.Content, "shutdown") {
				for _, channel := range config.Discord.Channels {
					if _, err := s.ChannelMessageSend(channel, fmt.Sprintf("***>>> Shutdown triggered by %s! (%s)***", m.Author.Username, m.Author.ID)); err != nil {
						log.Printf("[%s] Message error: %+v", m.ChannelID, err)
					}
				}
				logger.Printf("Shutdown triggered by %s! (%s)", m.Author.Username, m.Author.ID)
				cleanup()
				os.Exit(0)
			} else if IsCommand(m.Content, "stopall") {
				log.Printf("Warning: stopall triggered by %s! (%s)", m.Author.Username, m.Author.ID)
				for id, tailer := range Tails {
					tailer.Stop()
					tailer.Cleanup()
					delete(Onces, id)
					delete(Tails, id)
					delete(Servers, id)
					if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("***>>> Stopping game feed! (ID: %s)***", id)); err != nil {
						log.Printf("[%s] Message error: %+v", m.ChannelID, err)
					}
				}
			} else if IsCommand(m.Content, "stop") {
				_, _, args := GetCommand(m.Content)
				for id, tailer := range Tails {
					if strings.EqualFold(id, args) {
						tailer.Stop()
						tailer.Cleanup()
						delete(Onces, id)
						delete(Tails, id)
						delete(Servers, id)
						if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("***>>> Stopping game feed! (ID: %s)***", id)); err != nil {
							log.Printf("[%s] Message error: %+v", m.ChannelID, err)
						}
					}
				}
			} else if IsCommand(m.Content, "startall") {
				log.Printf("Warning: startall triggered by %s! (%s)", m.Author.Username, m.Author.ID)
				for _, settings := range config.Logs {
					go MessageParser(dg, settings)
				}
			} else if IsCommand(m.Content, "start") {
				_, _, args := GetCommand(m.Content)
				for _, settings := range config.Logs {
					if strings.EqualFold(settings.ID, args) {
						go MessageParser(dg, settings)
					}
				}
			} else if IsCommand(m.Content, "killfeed") {
				_, _, args := GetCommand(m.Content)
				for id, settings := range Servers {
					if strings.EqualFold(id, args) {
						settings.Killfeed = !settings.Killfeed
						if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("***>>> Killfeed for %s is now %s***", id, map[bool]string{
							true:  "ON",
							false: "OFF",
						}[settings.Killfeed])); err != nil {
							log.Printf("[%s] Message error: %+v", m.ChannelID, err)
						}
					}
				}
			}
		}
	})
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	if err = dg.Open(); err != nil {
		log.Fatal("error opening connection,", err)
		return
	}
	defer dg.Close()

	for _, settings := range config.Logs {
		IDs = append(IDs, settings.ID)
		if settings.OnStart {
			go MessageParser(dg, settings)
		}
	}

	for _, channel := range config.Discord.Channels {
		if _, err := dg.ChannelMessageSend(channel, fmt.Sprintf("***>>> %s v%s+%s is now online! (Server IDs: %s)***", Project, Version, Revision, strings.Join(IDs, ", "))); err != nil {
			log.Printf("[%s] Message error: %+v", channel, err)
		}
	}

	logger.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
