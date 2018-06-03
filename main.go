package main

import (
	"fmt"
	"github.com/lrstanley/girc"
	"log"
	"os"
	"strconv"
	"time"
	"strings"
	"github.com/xanzy/go-gitlab"
	"regexp"
)

type Config struct {
	Api ApiConfig
	Irc IrcConfig
}

type IrcConfig struct {
	Server       string
	Port         int
	Secure       bool
	Nick         string
	Ident        string
	Realname     string
	SaslAccount  string
	SaslPassword string
	SaslEnabled  bool
	Channels     []string
}

type ApiConfig struct {
	Url     string
	Key     string
	Project string
}

func NewConfigFromEnv() Config {
	var err error
	config := Config{}

	config.Irc.Server = os.Getenv("IRC_SERVER")
	config.Irc.Port, err = strconv.Atoi(os.Getenv("IRC_PORT"))
	if err != nil {
		panic(err)
	}
	config.Irc.Secure = os.Getenv("IRC_SECURE") == "true"
	config.Irc.Nick = os.Getenv("IRC_NICK")
	config.Irc.Ident = os.Getenv("IRC_IDENT")
	config.Irc.Realname = os.Getenv("IRC_REALNAME")
	config.Irc.SaslEnabled = os.Getenv("IRC_SASL_ENABLED") == "true"
	config.Irc.SaslAccount = os.Getenv("IRC_SASL_ACCOUNT")
	config.Irc.SaslPassword = os.Getenv("IRC_SASL_PASSWORD")

	config.Irc.Channels = strings.Split(os.Getenv("IRC_CHANNELS"), ",")

	config.Api.Url = os.Getenv("API_URL")
	config.Api.Key = os.Getenv("API_KEY")
	config.Api.Project = os.Getenv("API_PROJECT")

	return config
}

func main() {
	config := NewConfigFromEnv()

	ircConfig := girc.Config{
		Server: config.Irc.Server,
		Port:   config.Irc.Port,
		SSL:    config.Irc.Secure,
		Nick:   config.Irc.Nick,
		User:   config.Irc.Ident,
		Name:   config.Irc.Realname,
	}
	if config.Irc.SaslEnabled {
		ircConfig.SASL = &girc.SASLPlain{
			User: config.Irc.SaslAccount,
			Pass: config.Irc.SaslPassword,
		}
	}
	client := girc.New(ircConfig)

	api := gitlab.NewClient(nil, config.Api.Key)
	api.SetBaseURL(config.Api.Url)

	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		for _, name := range config.Irc.Channels {
			fmt.Printf("Joining %s\n", name)
			c.Cmd.Join(name)
		}
	})

	issueRegex := regexp.MustCompile("#\\d+\\b")
	client.Handlers.Add(girc.PRIVMSG, func(client *girc.Client, event girc.Event) {
		issues := issueRegex.FindAllString(event.Trailing, -1)
		for _, idString := range issues {
			id, err := strconv.Atoi(idString[1:])
			if err != nil {
				continue
			}
			issue, _, err := api.Issues.GetIssue(config.Api.Project, id)
			if err != nil {
				continue
			}
			if issue.ClosedAt != nil {
				client.Cmd.Notice(event.Source.Name, fmt.Sprintf("#%d (closed): %s – %s", issue.IID, issue.Title, issue.WebURL))
			} else {
				client.Cmd.Notice(event.Source.Name, fmt.Sprintf("#%d: %s – %s", issue.IID, issue.Title, issue.WebURL))
			}
		}
	})

	for {
		if err := client.Connect(); err != nil {
			log.Printf("error: %s", err)

			log.Println("reconnecting in 30 seconds...")
			time.Sleep(30 * time.Second)
		} else {
			return
		}
	}
}
