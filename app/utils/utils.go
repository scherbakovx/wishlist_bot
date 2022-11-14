package utils

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"log"
	"math/rand"
	"net/http"
	"regexp"

	"github.com/dyatlov/go-opengraph/opengraph"
)

func SeedRand() *rand.Rand {
	var b [8]byte
	_, err := crand.Read(b[:])
	if err != nil {
		log.Panic(err)
	}
	r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(b[:]))))
	return r
}

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
