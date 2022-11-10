package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

var startKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Writer", fmt.Sprint(models.Writer)),
		tgbotapi.NewInlineKeyboardButtonData("Reader", fmt.Sprint(models.Reader)),
	),
)

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
		case "start":
			return "Please select, are you gonna add your wished or read others?", nil
		case "get":
			if requestedChatId := user.ReadingUserId; requestedChatId != 0 {
				// TODO: I'm sure I don't need to make two requests here, but need more reading of GORM docs
				var requestedUser models.User
				result := database.Clauses(clause.OnConflict{DoNothing: true}).First(&requestedUser, "chat_id = ?", requestedChatId)

				if result.Error != nil {
					if result.Error.Error() == "record not found" {
						return fmt.Sprintf("No user with %v chat id", requestedChatId), nil
					} else {
						return botErrorMessage, result.Error
					}
				}

				var wish models.LocalWish
				result = database.Clauses(clause.OnConflict{DoNothing: true}).First(&wish, "user_id = ?", requestedUser.Id)
				if result.Error != nil {
					if result.Error.Error() == "record not found" {
						return fmt.Sprintf("User %v has no wishes :(", requestedChatId), nil
					} else {
						return botErrorMessage, result.Error
					}
				} else {
					return wish.String(), nil
				}
			}
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
						return "You have no wishes :(", nil
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
		if update.Message.Contact != nil {

			chatId := update.Message.Chat.ID

			msg := tgbotapi.NewMessage(chatId, "")
			msg.ReplyToMessageID = update.Message.MessageID

			var user models.User
			result := database.Clauses(clause.OnConflict{DoNothing: true}).First(&user, "chat_id = ?", update.Message.Contact.UserID)
			if result.Error != nil {
				if result.Error.Error() == "record not found" {
					msg.Text = "I don't know this user :("
				} else {
					log.Panic(result.Error)
				}
			} else {

				var currentUser models.User
				database.First(&currentUser, "chat_id = ?", chatId)
				database.Model(&currentUser).Update("reading_user_id", user.ChatId)

				msg.Text = "I know this user! Just tap /get and let's see what he/she wants :)"
			}

			bot.Send(msg)

		} else if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			chatId := update.Message.Chat.ID

			user, err := db.GetOrCreateUserInDB(database, chatId)
			if err != nil {
				log.Panic(err)
			}

			msg := tgbotapi.NewMessage(chatId, "")
			msg.ReplyToMessageID = update.Message.MessageID

			msg.Text, err = handleUserMessage(update, database, *user)
			if msg.Text == "Please select, are you gonna add your wished or read others?" {
				msg.ReplyMarkup = startKeyboard
			}
			if err != nil && msg.Text != botErrorMessage {
				log.Panic(err)
			}

			bot.Send(msg)
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Panic(err)
			}

			chatId := update.CallbackQuery.Message.Chat.ID

			user, err := db.GetOrCreateUserInDB(database, chatId)
			if err != nil {
				log.Panic(err)
			}
			intUserStatus, err := strconv.Atoi(update.CallbackQuery.Data)

			var answer string
			if intUserStatus == int(models.Writer) {
				answer = "Your status is Writer — just send me link and I'll add it to your wishlist!"
			} else {
				answer = "Your status is Reader — just send me contact card or nickname and I'll be ready to give you advice :)"
			}

			if err != nil {
				log.Panic(err)
			}
			user.Status = models.UserStatus(intUserStatus)
			database.Save(&user)

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, answer)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
	}
}
