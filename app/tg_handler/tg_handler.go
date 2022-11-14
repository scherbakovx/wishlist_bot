package tghandler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

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

const apiBaseURL string = "http://api:3000"

var regexLinkFinder *regexp.Regexp = utils.GetRegexpObject()

var startKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Writer", fmt.Sprint(models.Writer)),
		tgbotapi.NewInlineKeyboardButtonData("Reader", fmt.Sprint(models.Reader)),
	),
)

type UpdateHandler struct {
	Update        tgbotapi.Update
	Database      *gorm.DB
	Bot           *tgbotapi.BotAPI
	User          *models.User
	MessageToSend *tgbotapi.MessageConfig
}

func (uh UpdateHandler) HandleAnyMessage() {

	defer func() {
		if v := recover(); v != nil {
			log.Println(v)
			err, ok := v.(error)
			if ok {
				uh.MessageToSend.Text = err.Error()
			}
		}
		if uh.MessageToSend != nil {
			uh.Bot.Send(*uh.MessageToSend)
		}
	}()

	uh.constructMessageForAnswer()

	var err error
	uh.User, err = db.GetOrCreateUserInDB(uh.Database, uh.MessageToSend.ChatID)
	if err != nil {
		panic(err)
	}

	switch {
	case uh.Update.Message.Contact != nil:
		uh.handleContactMessage()

	case uh.Update.Message.Text != "":
		if link := regexLinkFinder.FindString(uh.Update.Message.Text); link != "" {
			uh.handleLinkMessage(link)

		} else if uh.Update.Message.IsCommand() {
			uh.handleCommandMessage()

		} else {
			uh.MessageToSend.Text = "I know only /get command"
		}

	case uh.Update.CallbackQuery != nil:
		uh.handleCallbackMessage()

	default:
		uh.MessageToSend.Text = "I don't know how to handle such message types."
	}
}

func (uh *UpdateHandler) constructMessageForAnswer() {

	var chatId int64
	var replyToMessageID int

	switch {
	case uh.Update.Message != nil:
		chatId = uh.Update.Message.Chat.ID
		replyToMessageID = uh.Update.Message.MessageID
	case uh.Update.CallbackQuery != nil:
		chatId = uh.Update.CallbackQuery.Message.Chat.ID
	}

	msg := tgbotapi.NewMessage(chatId, "")
	if replyToMessageID != 0 {
		msg.ReplyToMessageID = replyToMessageID
	}

	uh.MessageToSend = &msg
}

func (uh UpdateHandler) handleContactMessage() {

	var userToRead models.User
	result := uh.Database.Clauses(clause.OnConflict{DoNothing: true}).First(&userToRead, "chat_id = ?", uh.Update.Message.Contact.UserID)
	if result.Error != nil {
		panic(result.Error)
	}

	uh.Database.Model(&uh.User).Update("reading_user_id", userToRead.ChatId)
	uh.MessageToSend.Text = "I know this user! Just tap /get and let's see what he/she wants :)"
}

func (uh UpdateHandler) handleCallbackMessage() {
	callback := tgbotapi.NewCallback(uh.Update.CallbackQuery.ID, uh.Update.CallbackQuery.Data)
	if _, err := uh.Bot.Request(callback); err != nil {
		panic(err)
	}

	intUserStatus, err := strconv.Atoi(uh.Update.CallbackQuery.Data)
	if err != nil {
		panic(err)
	}

	if intUserStatus == int(models.Writer) {
		uh.MessageToSend.Text = "Your status is Writer — just send me link and I'll add it to your wishlist!"
	} else {
		uh.MessageToSend.Text = "Your status is Reader — just send me contact card or nickname and I'll be ready to give you advice :)"
	}

	uh.User.Status = models.UserStatus(intUserStatus)
	uh.Database.Save(&uh.User)
}

func (uh UpdateHandler) handleLinkMessage(link string) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprint(apiBaseURL, "/add_wish_to_user?chat_id=", uh.User.ChatId, "&link=", link), nil)
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		panic(res.Status)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	uh.MessageToSend.Text = string(b)
}

func (uh UpdateHandler) handleCommandMessage() {
	switch uh.Update.Message.Command() {
	case "start":
		uh.MessageToSend.Text = "Please select, are you gonna add your wished or read others?"
		uh.MessageToSend.ReplyMarkup = startKeyboard
	case "get":
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprint(apiBaseURL, "/get_user_wish?user_id=", uh.User.ChatId), nil)
		if err != nil {
			panic(err)
		}

		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			panic(res.Status)
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		uh.MessageToSend.Text = string(b)
	default:
		uh.MessageToSend.Text = "I know only /get command"
	}
}
