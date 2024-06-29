package main

import (
	"fmt"
	"os"
	"regexp"
	"ws-crawler/clawler"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <url>", os.Args[0])
		return
	}
	url := os.Args[1]
	if !isURL(url) {
		fmt.Printf("%s is not a valid url\n", url)
		fmt.Println("Hint: Punycode is escaped? or Prefix with https or http?")
		return
	}
	c := clawler.CreateClawler(url, false, os.Args[2:])

	c.Start()
}

var STATIC_REGEXP_URL, _ = regexp.Compile(`^https?://(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

func isURL(link string) bool {
	return STATIC_REGEXP_URL.MatchString(link)
}