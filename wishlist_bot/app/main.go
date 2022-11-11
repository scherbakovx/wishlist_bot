package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/scherbakovx/wishlist_bot/app/db"
	tghandler "github.com/scherbakovx/wishlist_bot/app/tg_handler"
	"gorm.io/gorm"
)

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

		updateHandler := tghandler.UpdateHandler{
			Update:   update,
			Database: database,
			Bot:      bot,
		}

		updateHandler.HandleAnyMessage()
	}
}
