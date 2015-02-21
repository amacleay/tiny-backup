package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const ARCHIVE_URL string = "http://www.npr.org/partials/music/series/tiny-desk-concerts/archive?start="

type ConcertUrlGroup []string

func (urls ConcertUrlGroup) IsEmpty() bool {
	return len(urls) == 0 ||
		urls[0] == ""
}

func main() {
	concert_channel := make(chan string, 10)
	go grabConcertUrls(concert_channel)
	for concert_url := range concert_channel {
		go ensure_concert_backed_up(concert_url)
	}
}

func grabConcertUrls(concert_channel chan string) {
	for i := 0; true; i++ {
		resp, _ := http.Get(fmt.Sprintf("%s%d", ARCHIVE_URL, 10*i))
		defer resp.Body.Close()
		body_bytes, _ := ioutil.ReadAll(resp.Body)
		content := string(body_bytes)

		concerts := ConcertUrls(content)
		if concerts.IsEmpty() {
			close(concert_channel)
			return
		}

		for concert_index := 0; concert_index < len(concerts); concert_index++ {
			concert_channel <- concerts[concert_index]
		}
	}
}

func ensure_concert_backed_up(url string) {
	fmt.Println(url)
}

func ConcertUrls(html string) ConcertUrlGroup {
	a := ConcertUrlGroup(make([]string, 10))
	concert_regex := regexp.MustCompile("http://www[.]npr[.]org/[^?\\s]*-concert")
	urls := concert_regex.FindAllString(html, 100)

	unique_urls := make(map[string]int)

	// Get the unique elements from the array
	for i := 0; i < len(urls); i++ {
		trimUrl := strings.TrimSpace(urls[i])
		if !strings.Contains(trimUrl, "series/tiny-desk") {
			unique_urls[trimUrl] = 1
		}
	}

	target_idx := 0
	for key, _ := range unique_urls {
		if len(key) > 0 {
			a[target_idx] = key
		}
		target_idx++
	}
	return a
}
