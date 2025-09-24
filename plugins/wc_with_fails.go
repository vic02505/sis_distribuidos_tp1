package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"tp1/mr"
)

func Map(filename string, content string) []mr.KeyValue {

	failProbability := rand.Float64()
	if failProbability >= 0 && failProbability <= 0.2 {
		log.Printf("I die x _ x")
		os.Exit(1)
	}

	words := strings.Fields(content)
	var wordCount []mr.KeyValue
	for _, word := range words {
		wordCount = append(wordCount, mr.KeyValue{Key: word, Value: "1"})
	}
	return wordCount
}

func Reduce(key string, values []string) string {

	failProbability := rand.Float64()
	if failProbability >= 0 && failProbability <= 0.2 {
		log.Printf("I die x _ x")
		os.Exit(1)
	}

	return strconv.Itoa(len(values))
}
