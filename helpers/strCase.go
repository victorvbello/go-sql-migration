package helpers

import (
	"regexp"
	"strings"
)

const regexCamelCase = " |-|_"

func StrCamelCase(str string, customRegex string) string {
	var key string
	strRegex := regexCamelCase
	if customRegex != "" {
		strRegex = customRegex
	}
	re := regexp.MustCompile(strRegex)
	words := re.Split(str, -1)
	for _, word := range words {
		key += strings.Title(word)
	}
	return key
}

func StrSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])()")
	matchOthers := regexp.MustCompile(`[\.\s]`)
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	snake = matchOthers.ReplaceAllString(snake, "_")
	return strings.ToLower(snake)
}
