package clawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

type Clawler struct {
	URL                 string
	FoundRoutes         []string
	CurrentSearchRoutes []string
	SameOrigin          bool
}

func CreateClawler(url string, sameOrigin bool) *Clawler {
	return &Clawler{
		URL:                 url,
		FoundRoutes:         []string{url},
		CurrentSearchRoutes: []string{url},
		SameOrigin:          sameOrigin,
	}
}

func (c *Clawler) Start() {
	fmt.Printf("Crawling %s\n", c.URL)
	c.Clawl()
}

func (c *Clawler) Clawl() {
	for {
		route := c.CurrentSearchRoutes[0]
		c.CurrentSearchRoutes = c.CurrentSearchRoutes[1:]

		FoundResult := FindLinks(route, c.SameOrigin, c.URL)

		for _, result := range removeMultipleValue(FoundResult, c.FoundRoutes...) {
			fmt.Printf("Found: %s\n", result)
		}

		c.CurrentSearchRoutes = removeMultipleValue(append(c.CurrentSearchRoutes, FoundResult...), c.FoundRoutes...)
		c.FoundRoutes = removeMultipleValue(append(c.FoundRoutes, FoundResult...))

		if len(c.CurrentSearchRoutes) == 0 {
			fmt.Printf("Crawling %s done\n", c.URL)
			break
		}
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

func FindLinks(url string, sameOrigin bool, originUrl string) []string {
	fmt.Println(url)
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
		return sanitizeLinks(foundLinks, url, sameOrigin, originUrl)
	}

	return []string{}
}

var STATIC_REGEXP_URL, _ = regexp.Compile(`^https?://(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
var STATIC_REGEXP_NO_PREFIX_URL, _ = regexp.Compile(`^[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

func isAbsoluteURL(link string) bool {
	return STATIC_REGEXP_URL.MatchString(link)
}

func isHostname(link string) bool {
	return STATIC_REGEXP_NO_PREFIX_URL.MatchString(link)
}

var STATIC_REGEXP_RELATIVE_PATH, _ = regexp.Compile(`^(\.|\.\.)\/.*`)

func isRelativePath(link string) bool {
	return STATIC_REGEXP_RELATIVE_PATH.MatchString(link)
}

func sanitizeLinks(links []string, baseURL string, sameOrigin bool, originUrl string) []string {
	result := []string{}

	for _, link := range links {
		var url string

		if isAbsoluteURL(link) {
			url = link
		} else if isHostname(link) {
			url = "https://"+link
		} else if isRelativePath(link) {
			url = urlFixer(baseURL)+"/"+pathFixer(link)
		} else {
			url = "https://"+urlFixer(urlPickHost(baseURL))+"/"+pathFixer(link)
		}

		if sameOrigin && !strings.HasPrefix(url, originUrl) {
			continue
		}
		
		result = append(result, url)
	}

	return result
}

func urlFixer(url string) string {
	if strings.HasSuffix(url, "/") {
		return url[:len(url)-1]
	}
	return url
}

func pathFixer(url string) string {
	if strings.HasPrefix(url, "/") {
		return url[1:]
	}
	return url
}

func urlPickHost(url string) string {
	if isAbsoluteURL(url) {
		url = strings.Split(url, "/")[2]
	} else {
		url = strings.Split(url, "/")[0]
	}

	return url
}
