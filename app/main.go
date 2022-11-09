package main

import (
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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

var randomizer *rand.Rand = utils.SeedRand()
var regexLinkFinder *regexp.Regexp = utils.GetRegexpObject()

const botErrorMessage string = "I'm broken"
const botSuccessfulMessage string = "Added!"

func handleUserMessage(update tgbotapi.Update, database *gorm.DB, user models.User) (string, error) {

	var answer string
	var err error

	if link := regexLinkFinder.FindString(update.Message.Text); link != "" {
		if user.AirTable.Board != "" {
			err = airtable.InsertDataToAirTable(client, link)
			if err != nil {
				return botErrorMessage, err
			}
		} else {
			openGraphData, _ := utils.GetOGTags(client, link)
			wish := models.LocalWish{
				Wish: models.Wish{
					Name: openGraphData.Title,
					Link: openGraphData.URL,
				},
				UserId: user.Id,
			}
			result := database.Create(&wish)

			if result.Error != nil {
				return botErrorMessage, result.Error
			}
		}
		return botSuccessfulMessage, nil
	} else if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "get":
			if user.AirTable.Board != "" {
				randomObjectData, err := airtable.GetDataFromAirTable(client, randomizer)
				if err != nil {
					return botErrorMessage, err
				}
				return randomObjectData, nil
			} else {
				var wish models.LocalWish
				result := database.Clauses(clause.OnConflict{DoNothing: true}).First(&wish)

				if result.Error != nil {
					if result.Error.Error() == "record not found" {
						return "User has no wishes :(", nil
					} else {
						return botErrorMessage, result.Error
					}
				} else {
					return wish.String(), nil
				}
			}
		default:
			answer = "I know only /get command"
		}
	} else {
		answer = "I know only /get command"
	}

	return answer, err
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Panic("failed to open .env file")
	}
	var database *gorm.DB = db.Init()

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

		chatId := update.Message.Chat.ID

		user, err := db.GetOrCreateUserInDB(database, chatId)
		if err != nil {
			log.Panic(err)
		}

		msg := tgbotapi.NewMessage(chatId, "")
		msg.ReplyToMessageID = update.Message.MessageID

		msg.Text, err = handleUserMessage(update, database, *user)
		if err != nil && msg.Text != botErrorMessage {
			log.Panic(err)
		}

		bot.Send(msg)
	}
}
