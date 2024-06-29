package main

import (
	"ws-crawler/clawler"
)

func main() {
	c := clawler.CreateClawler("https://activetk.jp")

	c.Start()
}

