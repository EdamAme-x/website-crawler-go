package clawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"golang.org/x/exp/slices"
)

type Clawler struct {
	URL                 string
	FoundRoutes         []string
	CurrentSearchRoutes []string
}

func CreateClawler(url string) *Clawler {
	return &Clawler{
		URL:                 url,
		FoundRoutes:         []string{url},
		CurrentSearchRoutes: []string{url},
	}
}

func (c *Clawler) Start() {
	fmt.Printf("Crawling %s\n", c.URL)
	c.Clawl()
}

func (c *Clawler) Clawl() {
	for len(c.CurrentSearchRoutes) > 0 {
		route := c.CurrentSearchRoutes[0]
		c.CurrentSearchRoutes = c.CurrentSearchRoutes[1:]

		FoundResult := FindLinks(route)

		for _, result := range removeMultipleValue(FoundResult, c.FoundRoutes...) {
			fmt.Printf("Found: %s\n", result)
		}

		c.CurrentSearchRoutes = removeMultipleValue(append(c.CurrentSearchRoutes, FoundResult...), c.FoundRoutes...)
		c.FoundRoutes = removeMultipleValue(append(c.FoundRoutes, FoundResult...))
	}
}

func removeMultipleValue(values []string, subValues ...string) []string {
	result := []string{}

	for _, value := range values {
		if !slices.Contains(append(subValues, result...), value) {
			result = append(result, value)
		}
	}

	return result
}

var STATIC_REGEXP_HREF_AND_SRC, _ = regexp.Compile(`\s(href|src)=["'](.+?)['"][\s>]`)
var STATIC_REGEXP_WINDOW_OPEN, _ = regexp.Compile(`window\.open\(["'](.+?)["']`)

func FindLinks(url string) []string {

	resp, err := http.Get(url)

	if err != nil {
		return []string{}
	}

	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	body := string(byteArray)

	links := STATIC_REGEXP_HREF_AND_SRC.FindAllStringSubmatch(body, -1)

	foundLinks := []string{}

	for _, link := range links {
		foundLinks = append(foundLinks, link[2])
	}

	links = STATIC_REGEXP_WINDOW_OPEN.FindAllStringSubmatch(body, -1)

	for _, link := range links {
		foundLinks = append(foundLinks, link[1])
	}

	if len(foundLinks) > 0 {
		return sanitizeLinks(foundLinks)
	}

	return []string{}
}

func sanitizeLinks(links []string) []string {
	return links
}