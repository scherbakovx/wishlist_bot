package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/scherbakovx/wishlist_bot/app/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

var randomizer = utils.SeedRand()

type AirTableObjectFields struct {
	Name  string `json:"Name"`
	Price int    `json:"Price ($)"`
	Link  string `json:"Link"`
}

type AirTableSingleObject struct {
	Id     string               `json:"id"`
	Fields AirTableObjectFields `json:"fields"`
}

type AirTableObjectsArray struct {
	Records []AirTableSingleObject `json:"records"`
}

func getDataFromAirTable() string {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.airtable.com/v0/appEXUeaG06r5KYBe/Wishlist", nil)
	if err != nil {
		log.Panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("AIRTABLE_TOKEN")))
	res, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unexpected status: got %v", res.Status))
	}
	var objects AirTableObjectsArray
	err = json.NewDecoder(res.Body).Decode(&objects)
	if err != nil {
		log.Panic(err)
	}

	randomIndex := randomizer.Intn(len(objects.Records))

	objectFromWishlist := objects.Records[randomIndex]
	result := fmt.Sprintf("Name: %s\nPrice: %d\nLink: %s", objectFromWishlist.Fields.Name, objectFromWishlist.Fields.Price, objectFromWishlist.Fields.Link)
	return result
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Panic("failed to open .env file")
	}

	bot_token := os.Getenv("BOT_TOKEN")

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

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "get":
				randomObjectData := getDataFromAirTable()
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
