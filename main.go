package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const (
	AppTokenVar    = "APP_TOKEN"
	BotTokenVar    = "BOT_TOKEN"
	VerifyTokenVar = "VERIFY_TOKEN"

	ErrEmptyOrUnsetTemplate = "%s environment variable is unset or empty"
)

type Config struct {
	AppToken    string
	BotToken    string
	VerifyToken string
}

func NewConfigFromEnviron() (*Config, error) {
	config := &Config{}

	config.AppToken = os.Getenv(AppTokenVar)
	if config.AppToken == "" {
		return nil, fmt.Errorf(ErrEmptyOrUnsetTemplate, AppTokenVar)
	}
	if !strings.HasPrefix(config.AppToken, "xapp") {
		return nil, errors.New("provided app token doesn't start with xapp")
	}

	config.BotToken = os.Getenv(BotTokenVar)
	if config.BotToken == "" {
		return nil, fmt.Errorf(ErrEmptyOrUnsetTemplate, BotTokenVar)
	}
	if !strings.HasPrefix(config.BotToken, "xoxb") {
		return nil, errors.New("provided bot token doesn't start with xoxb")
	}

	config.VerifyToken = os.Getenv(VerifyTokenVar)
	if config.VerifyToken == "" {
		return nil, fmt.Errorf(ErrEmptyOrUnsetTemplate, VerifyTokenVar)
	}

	return config, nil
}

func main() {
	config, err := NewConfigFromEnviron()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	client := socketmode.New(
		slack.New(
			config.BotToken,
			slack.OptionAppLevelToken(config.AppToken),
		),
	)

	go func() {
		for event := range client.Events {
			command, ok := event.Data.(slack.SlashCommand)
			if !ok {
				continue
			}

			if !command.ValidateToken(config.VerifyToken) {
				continue
			}

			payload := map[string]interface{}{
				"response_type": slack.ResponseTypeInChannel,
				"blocks": []slack.Block{
					slack.SectionBlock{
						Type: slack.MBTSection,
						Text: &slack.TextBlockObject{
							Type: slack.MarkdownType,
							Text: fmt.Sprintf(
								"<@%s> has lost the donut game! :doughnut:",
								command.UserID,
							),
						},
					},
				},
			}

			client.Ack(*event.Request, payload)
		}
	}()

	client.Run()
}
