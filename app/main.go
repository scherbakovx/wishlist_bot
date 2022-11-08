package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/scherbakovx/wishlist_bot/app/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

var randomizer = utils.SeedRand()

type AirTableImageObject struct {
	Url string `json:"url"`
}

type AirTableObjectFields struct {
	Name  string                `json:"Name"`
	Price int                   `json:"Price ($)"`
	Link  string                `json:"Link"`
	Image []AirTableImageObject `json:"Image"`
}

type AirTableSingleObject struct {
	Id     string               `json:"id,omitempty"`
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

func insertDataToAirTable(link string) {

	newObject := AirTableObjectsArray{
		Records: []AirTableSingleObject{
			{
				Fields: AirTableObjectFields{
					Link: link,
				},
			},
		},
	}
	btResult, _ := json.Marshal(&newObject)

	fmt.Println(bytes.NewBuffer(btResult))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.airtable.com/v0/appEXUeaG06r5KYBe/Wishlist", bytes.NewBuffer(btResult))
	if err != nil {
		log.Panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("AIRTABLE_TOKEN")))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unexpected status: got %v", res.Status))
	}
}

func getLinkFromMessage(message string) string {
	r, _ := regexp.Compile(`(?:(?:https?|ftp):\/\/)?[\w/\-?=%.]+\.[\w/\-&?=%.]+`)
	return r.FindString(message)
}

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

		if link := getLinkFromMessage(update.Message.Text); link != "" {
			fmt.Println("Chat ID", update.Message.Chat.ID)
			if update.Message.Chat.ID == 16803083 {
				insertDataToAirTable(link)
				msg.Text = fmt.Sprint("I've added this: ", link)
			} else {
				msg.Text = "Sorry, only Anton could add links to his Wishlist"
			}
		} else if update.Message.IsCommand() {
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
