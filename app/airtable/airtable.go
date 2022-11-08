package airtable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
)

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

func GetDataFromAirTable(client *http.Client, randomizer *rand.Rand) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.airtable.com/v0/appEXUeaG06r5KYBe/Wishlist", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("AIRTABLE_TOKEN")))
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: got %v", res.Status)
	}
	var objects AirTableObjectsArray
	err = json.NewDecoder(res.Body).Decode(&objects)
	if err != nil {
		return "", err
	}

	randomIndex := randomizer.Intn(len(objects.Records))

	objectFromWishlist := objects.Records[randomIndex]
	result := fmt.Sprintf("Name: %s\nPrice: %d$\nLink: %s", objectFromWishlist.Fields.Name, objectFromWishlist.Fields.Price, objectFromWishlist.Fields.Link)
	return result, nil
}

func InsertDataToAirTable(client *http.Client, link string) error {

	newWish := AirTableObjectsArray{
		Records: []AirTableSingleObject{
			{
				Fields: AirTableObjectFields{
					Link: link,
				},
			},
		},
	}
	marshaledNewWish, _ := json.Marshal(&newWish)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.airtable.com/v0/appEXUeaG06r5KYBe/Wishlist", bytes.NewBuffer(marshaledNewWish))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("AIRTABLE_TOKEN")))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: got %v", res.Status)
	}

	return nil
}
