package tghandler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/scherbakovx/wishlist_bot/app/airtable"
	"github.com/scherbakovx/wishlist_bot/app/db"
	"github.com/scherbakovx/wishlist_bot/app/models"
	"github.com/scherbakovx/wishlist_bot/app/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

var regexLinkFinder *regexp.Regexp = utils.GetRegexpObject()
var randomizer *rand.Rand = utils.SeedRand()

const botErrorMessage string = "I'm broken"
const botSuccessfulMessage string = "Added!"

var startKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Writer", fmt.Sprint(models.Writer)),
		tgbotapi.NewInlineKeyboardButtonData("Reader", fmt.Sprint(models.Reader)),
	),
)

type UpdateHandler struct {
	Update   tgbotapi.Update
	Database *gorm.DB
	Bot      *tgbotapi.BotAPI
	User     *models.User
}

func (uh UpdateHandler) HandleAnyMessage() {
	var messageToSend *tgbotapi.MessageConfig

	if uh.Update.Message.Contact != nil {
		messageToSend = uh.handleContactMessage()

	} else if uh.Update.Message != nil {
		messageToSend = uh.handleTextMessage()

	} else if uh.Update.CallbackQuery != nil {
		messageToSend = uh.handleCallbackMessage()
	}

	if messageToSend != nil {
		uh.Bot.Send(*messageToSend)
	}
}

func (uh UpdateHandler) handleContactMessage() *tgbotapi.MessageConfig {
	chatId := uh.Update.Message.Chat.ID

	msg := tgbotapi.NewMessage(chatId, "")
	msg.ReplyToMessageID = uh.Update.Message.MessageID

	var user models.User
	result := uh.Database.Clauses(clause.OnConflict{DoNothing: true}).First(&user, "chat_id = ?", uh.Update.Message.Contact.UserID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			msg.Text = "I don't know this user :("
		} else {
			log.Panic(result.Error)
		}
	} else {

		var currentUser models.User
		uh.Database.First(&currentUser, "chat_id = ?", chatId)
		uh.Database.Model(&currentUser).Update("reading_user_id", user.ChatId)

		msg.Text = "I know this user! Just tap /get and let's see what he/she wants :)"
	}

	return &msg
}

func (uh UpdateHandler) handleTextMessage() *tgbotapi.MessageConfig {

	var err error

	msg := tgbotapi.NewMessage(uh.User.ChatId, "")
	msg.ReplyToMessageID = uh.Update.Message.MessageID

	msg.Text, err = uh.handleUserMessage()
	if msg.Text == "Please select, are you gonna add your wished or read others?" {
		msg.ReplyMarkup = startKeyboard
	}
	if err != nil && msg.Text != botErrorMessage {
		log.Panic(err)
	}

	return &msg
}

func (uh UpdateHandler) handleCallbackMessage() *tgbotapi.MessageConfig {
	callback := tgbotapi.NewCallback(uh.Update.CallbackQuery.ID, uh.Update.CallbackQuery.Data)
	if _, err := uh.Bot.Request(callback); err != nil {
		log.Panic(err)
	}

	chatId := uh.Update.CallbackQuery.Message.Chat.ID

	user, err := db.GetOrCreateUserInDB(uh.Database, chatId)
	if err != nil {
		log.Panic(err)
	}
	intUserStatus, err := strconv.Atoi(uh.Update.CallbackQuery.Data)

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
	uh.Database.Save(&user)

	msg := tgbotapi.NewMessage(uh.Update.CallbackQuery.Message.Chat.ID, answer)
	if _, err := uh.Bot.Send(msg); err != nil {
		log.Panic(err)
	}

	return &msg
}

func (uh UpdateHandler) handleUserMessage() (string, error) {

	if link := regexLinkFinder.FindString(uh.Update.Message.Text); link != "" {
		return uh.handleLinkMessage(link)

	} else if uh.Update.Message.IsCommand() {
		return uh.handleCommandMessage()

	} else {
		return "I know only /get command", nil
	}
}

func (uh UpdateHandler) handleLinkMessage(link string) (string, error) {
	if uh.User.AirTable.Board != "" {
		err := airtable.InsertDataToAirTable(client, link)
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
			UserId: uh.User.Id,
		}
		result := uh.Database.Create(&wish)

		if result.Error != nil {
			return botErrorMessage, result.Error
		}
	}
	return botSuccessfulMessage, nil
}

func (uh UpdateHandler) handleCommandMessage() (string, error) {
	switch uh.Update.Message.Command() {
	case "start":
		return "Please select, are you gonna add your wished or read others?", nil
	case "get":
		if requestedChatId := uh.User.ReadingUserId; requestedChatId != 0 {

			var wish models.LocalWish
			result := uh.Database.Clauses(clause.OnConflict{DoNothing: true}).Model(&models.LocalWish{}).Joins("JOIN users ON local_wishes.user_id = users.id").Where("users.chat_id = ?", requestedChatId).First(&wish)
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
		if uh.User.AirTable.Board != "" {
			randomObjectData, err := airtable.GetDataFromAirTable(client, randomizer)
			if err != nil {
				return botErrorMessage, err
			}
			return randomObjectData, nil
		} else {
			var wish models.LocalWish
			result := uh.Database.Clauses(clause.OnConflict{DoNothing: true}).First(&wish)

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
		return "I know only /get command", nil
	}
}
