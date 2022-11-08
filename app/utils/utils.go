package utils

import (
	crand "crypto/rand"
	"encoding/binary"
	"log"
	"math/rand"
	"regexp"
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
