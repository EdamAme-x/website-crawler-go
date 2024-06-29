package main

import (
	"ws-crawler/clawler"
)

func main() {
	c := clawler.CreateClawler("https://activetk.jp", false)

	c.Start()
}

