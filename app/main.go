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
	"github.com/scherbakovx/wishlist_bot/app/db"
	"github.com/scherbakovx/wishlist_bot/app/models"
	"github.com/scherbakovx/wishlist_bot/app/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gorm.io/gorm/clause"
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

	DB := db.Init()

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID

		user := models.User{ChatId: update.Message.Chat.ID}
		result := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&user)

		if result.Error != nil {
			log.Panic(result.Error)
		}

		if result.RowsAffected == 0 {
			DB.First(&user, "chat_id = ?", user.ChatId)
		}

		if link := regexLinkFinder.FindString(update.Message.Text); link != "" {
			if user.AirTable.Board != "" {
				err = airtable.InsertDataToAirTable(client, link)
				if err != nil {
					msg.Text = "I'm broken"
				}
				msg.Text = fmt.Sprint("I've added this: ", link)
			} else {
				openGraphData, _ := utils.GetOGTags(client, link)
				wish := models.Wish{Name: openGraphData.Title, Link: openGraphData.URL, UserId: user.Id}
				result := DB.Create(&wish)

				if result.Error != nil {
					log.Panic(result.Error)
				}

				msg.Text = "Added to bot local DB!"
			}
		} else if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "get":
				if user.AirTable.Board != "" {
					randomObjectData, err := airtable.GetDataFromAirTable(client, randomizer)
					if err != nil {
						msg.Text = "I'm broken"
					}
					msg.Text = randomObjectData
				} else {
					var objectData models.Wish
					result := DB.First(&objectData)

					if result.Error != nil {
						log.Panic(result.Error)
					}

					msg.Text = fmt.Sprintf("Name: %s\nPrice: %d$\nLink: %s", objectData.Name, objectData.Price, objectData.Link)
				}

			default:
				msg.Text = "I don't know this command"
			}
		} else {
			msg.Text = "I know only /get command"
		}

		bot.Send(msg)

	}
}
