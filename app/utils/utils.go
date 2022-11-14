package utils

import (
	"context"
	"log"
	"net/http"
	"regexp"

	"github.com/dyatlov/go-opengraph/opengraph"
)

func GetRegexpObject() *regexp.Regexp {
	r, err := regexp.Compile(`(?:(?:https?|ftp):\/\/)?[\w/\-?=%.]+\.[\w/\-&?=%.]+`)
	if err != nil {
		log.Panic(err)
	}
	return r
}

func GetOGTags(client *http.Client, url string) (opengraph.OpenGraph, error) {

	var data opengraph.OpenGraph = *opengraph.NewOpenGraph()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return data, err
	}
	res, err := client.Do(req)
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	data.ProcessHTML(res.Body)

	return data, nil
}
