package clawler

import (
	"fmt"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type Clawler struct {
	URL                 string
	FoundRoutes         []string
	CurrentSearchRoutes []string
	SameOrigin          bool
	BlockList           []string
}

func CreateClawler(url string, sameOrigin bool, blockList []string) *Clawler {
	return &Clawler{
		URL:                 url,
		FoundRoutes:         []string{url},
		CurrentSearchRoutes: []string{url},
		SameOrigin:          sameOrigin,
		BlockList:           blockList,
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

		FoundResult := FindLinks(route, c.SameOrigin, c.URL, pickOriginFromURL(route), c.BlockList)

		for _, result := range removeMultipleValue(FoundResult, append(c.FoundRoutes, c.CurrentSearchRoutes...)...) {
			fmt.Printf("Found: %s\n", result)
		}

		c.CurrentSearchRoutes = removeMultipleValue(append(c.CurrentSearchRoutes, FoundResult...), c.FoundRoutes...)
		c.FoundRoutes = removeMultipleValue(append(c.FoundRoutes, route))

		if len(c.CurrentSearchRoutes) == 0 {
			fmt.Printf("Crawling %s done\n", c.URL)
			break
		}
	}
}

func pickOriginFromURL(url string) string {
	host := strings.Split(url, "//")
	if len(host) > 2 {
		return "https://" + strings.Split(url, "//")[1]
	}
	return url
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

var STATIC_REGEXP_HREF_AND_SRC, _ = regexp.Compile(`\s(href|src)=["'](.+?)['"][\s>/]`)
var STATIC_REGEXP_WINDOW_OPEN, _ = regexp.Compile(`window\.open\(["'](.+?)["']`)

func FindLinks(url string, sameOrigin bool, originUrl string, baseURL string, blockList []string) []string {
	fmt.Printf("Find: %s\n", url)
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

	foundLinks = removeMultipleValue(foundLinks)

	if len(foundLinks) > 0 {
		return sanitizeLinks(foundLinks, baseURL, sameOrigin, originUrl, url, blockList)
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

var STATIC_REGEXP_HASH = regexp.MustCompile(`^#.*`)

func isHash(link string) bool {
	return STATIC_REGEXP_HASH.MatchString(link)
}

func sanitizeLinks(links []string, baseURL string, sameOrigin bool, originUrl string, rawURL string, blockList []string) []string {
	result := []string{}

	for _, link := range links {
		var url string

		if strings.HasPrefix(link, "data:") {
			continue
		} else if strings.HasPrefix(link, "//") {
			link = "https:" + link
		}

		if isAbsoluteURL(link) {
			url = link
		} else if isHostname(link) {
			url = "https://" + urlFixer(link) + "/" + strings.Replace(rawURL, urlFixer(baseURL) + "/", "", 1)
		} else if isRelativePath(link) {
			url = urlFixer(rawURL) + "/" + pathFixer(link)
		} else if isHash(link) {
			url = urlFixer(rawURL) + "#" + link
		} else {
			url = "https://" + urlFixer(urlPickHost(baseURL)) + "/" + pathFixer(link)
		}

		url, err := fixURL(url, baseURL)

		isBlock := false

		for _, block := range blockList {
			if strings.Contains(url, block) {
				isBlock = true
				break
			}
		}

		if err != nil || sameOrigin && !strings.HasPrefix(url, originUrl) || isBlock {
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

var STATIC_REGEXP_INVALID_URL, _ = regexp.Compile(`^https?://(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9]{1,6}/(https?://.+)`)
var STATIC_REGEXP_JAVASCRIPT, _ = regexp.Compile(`.*[a-zA-Z0-9]{1,6}/javascript:.*`)

func fixURL(url string, baseURL string) (string, error) {
	if STATIC_REGEXP_JAVASCRIPT.MatchString(url) {
		return url, fmt.Errorf("JavaScript URL: %s", url)
	}

	url = strings.Trim(url, " ")
	urlSubmatch := STATIC_REGEXP_INVALID_URL.FindAllStringSubmatch(url, 1)

	if len(urlSubmatch) > 0 {
		url = urlSubmatch[0][1]
	}
	url = strings.Replace(url, urlFixer(baseURL)+"/https://", "https://", 1)
	url = strings.Replace(url, urlFixer(baseURL)+"/http://", "http://", 1)
	url = strings.ReplaceAll(url, "https://https:/", "https://")
	url = strings.ReplaceAll(url, "https://https://", "https://")

	return url, nil
}
