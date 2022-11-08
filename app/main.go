package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/scherbakovx/wishlist_bot/app/airtable"
	"github.com/scherbakovx/wishlist_bot/app/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

var randomizer *rand.Rand = utils.SeedRand()
var regexLinkFinder *regexp.Regexp = utils.GetRegexpObject()

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Panic("failed to open .env file")
	}

	bot_token := os.Getenv("TG_BOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(bot_token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID

		if link := regexLinkFinder.FindString(update.Message.Text); link != "" {
			fmt.Println("Chat ID", update.Message.Chat.ID)
			if update.Message.Chat.ID == 16803083 {
				err = airtable.InsertDataToAirTable(client, link)
				if err != nil {
					msg.Text = "I'm broken"
				}
				msg.Text = fmt.Sprint("I've added this: ", link)
			} else {
				msg.Text = "Sorry, only Anton could add links to his Wishlist"
			}
		} else if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "get":
				randomObjectData, err := airtable.GetDataFromAirTable(client, randomizer)
				if err != nil {
					msg.Text = "I'm broken"
				}
				msg.Text = randomObjectData
			default:
				msg.Text = "I don't know this command"
			}
		} else {
			msg.Text = "I know only /get command"
		}

		bot.Send(msg)

	}
}
