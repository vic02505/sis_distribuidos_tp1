package main

import (
    "strings"
    "strconv"
    "tp1/mr" 
)


func Map(content string) []mr.KeyValue{ 
	words := strings.Fields(content)
	var wordCount []mr.KeyValue
	for _, word := range words {
		wordCount = append(wordCount, mr.KeyValue{Key: word, Value: "1"})
	}
	return wordCount
}

func Reduce(key string, values []string) string {
	return strconv.Itoa(len(values)) 
}
