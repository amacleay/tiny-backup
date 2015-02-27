package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

const ARCHIVE_URL string = "http://www.npr.org/partials/music/series/tiny-desk-concerts/archive?start="
const CONCURRENT_URLS int = 10
const CONCURRENT_DOWNLOADS int = 50

type ConcertUrlGroup []string

func (urls ConcertUrlGroup) IsEmpty() bool {
	return len(urls) == 0 ||
		urls[0] == ""
}

func main() {
	concert_channel := make(chan string, CONCURRENT_URLS)
	go grabConcertUrls(concert_channel)

	concurrencyLimiter := make(chan bool, CONCURRENT_DOWNLOADS)
	for concert_url := range concert_channel {
		if len(concert_url) > 1 {
			select {
			case concurrencyLimiter <- true:
				// once the download is completed, pull one off concurrencyLimiter
				go ensure_concert_backed_up(concert_url, concurrencyLimiter)
			}
		}
	}
}

func grabConcertUrls(concert_channel chan string) {
	for i := 0; true; i++ {
		html_bytes, _ := getUrlBody(fmt.Sprintf("%s%d", ARCHIVE_URL, 10*i))
		content := string(html_bytes)

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

func getUrlBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body_bytes, err := ioutil.ReadAll(resp.Body)
	return body_bytes, err
}

func ensure_concert_backed_up(url string, concurrencyLimiter chan bool) {
	html_bytes, _ := getUrlBody(url)
	content := string(html_bytes)

	mp3_regex := regexp.MustCompile("http://[^?\\s]*[.]mp3")

	download_url := mp3_regex.FindString(content)
	if len(download_url) == 0 {
		fmt.Fprintf(os.Stderr, "BAD: %s\n", url)
	} else {
		base_name := path.Base(download_url)

		// Only download if the file doesn't exist already
		if stat, err := os.Stat(base_name); os.IsNotExist(err) {
			contents, err := getUrlBody(download_url)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to download %s: %v\n", download_url, err)
			} else {
				fmt.Printf("Download succeeded for %s\n", download_url)
				newFile, err := os.Create(base_name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to create %s\n", base_name)
				} else {
					defer newFile.Close()
					fmt.Printf("Writing to %s\n", base_name)
					_, err := newFile.Write(contents)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Write error: %s\n", base_name)
					}
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "File %s wasn't IsNotExist error: %v %v\n", base_name, stat, err)
		}
	}
	<-concurrencyLimiter
}

func ConcertUrls(html string) ConcertUrlGroup {
	a := ConcertUrlGroup(make([]string, 10)) // 10 urls per page max
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
