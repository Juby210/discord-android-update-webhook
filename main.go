package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Config .
type Config struct {
	Webhook          string
	Minutes          int
	SendIfEmptyCache bool
}

// Cache .
type Cache struct {
	LastVersion string
}

var config Config
var cache Cache

func interval(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func formatTime(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

func log(msg string) {
	fmt.Println("[" + formatTime(time.Now()) + "] [DAUW] " + msg)
}

func check() {
	log("Checking for new update..")

	client := &http.Client{}
	r, _ := http.NewRequest("GET", "https://www.apkmirror.com/uploads/?q=discord-chat-for-gamers", nil)
	r.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36 OPR/60.0.3255.170")
	resp, err := client.Do(r)
	if err != nil {
		log("[ERROR] " + err.Error())
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		log("[ERROR] " + err.Error())
	} else {
		appVer := strings.Replace(doc.Find(".infoslide-value").First().Text(), " ", "", -1)

		if cache.LastVersion != appVer {
			if cache.LastVersion != "" || config.SendIfEmptyCache {
				bd := map[string]interface{}{"embeds": []map[string]interface{}{map[string]interface{}{
					"author": map[string]string{
						"name":     strings.Replace(doc.Find("a.fontBlack").First().Text(), doc.Find(".infoslide-value").First().Text(), "", 1),
						"icon_url": "https://lh3.googleusercontent.com/_4zBNFjA8S9yjNB_ONwqBvxTvyXYdC7Nh1jYZ2x6YEcldBr2fyijdjM2J5EoVdTpnkA=s180-rw",
						"url":      "https://play.google.com/store/apps/details?id=com.discord&hl=en",
					},
					"title":       "New version: **" + appVer + "**",
					"description": "Updated: " + doc.Find("span.datetime_utc").First().Text(),
					"color":       7506394,
				}}}
				body, _ := json.Marshal(bd)
				log("Found new version " + appVer + ", Sending")
				http.Post(config.Webhook, "application/json", bytes.NewReader(body))
			}

			cache.LastVersion = appVer
			b, _ := json.Marshal(cache)
			ioutil.WriteFile("cache.json", b, 0644)
		}
	}
}

func main() {
	jfile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
		if os.IsNotExist(err) {
			return
		}
	}
	defer jfile.Close()
	byteV, _ := ioutil.ReadAll(jfile)
	json.Unmarshal(byteV, &config)

	jfile, err = os.Open("cache.json")
	if err != nil {
		if os.IsNotExist(err) == false {
			fmt.Println(err)
		}
	}
	defer jfile.Close()
	byteV, _ = ioutil.ReadAll(jfile)
	json.Unmarshal(byteV, &cache)

	log("Started | Interval: " + fmt.Sprint(config.Minutes) + " minute/s")

	<-interval(check, time.Duration(config.Minutes)*time.Minute)
}
