package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	scraper "github.com/juby210-PL/google-play-scraper"
)

// Config .
type Config struct {
	Webhook          string
	Minutes          int
	SendIfEmptyCache bool
}

// Cache .
type Cache struct {
	LastVersion   string
	LastChangelog string
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
	app, err := scraper.GetApp("com.discord")
	if err != nil {
		log("[ERROR] " + err.Error())
	} else {
		if cache.LastVersion != "" || config.SendIfEmptyCache {
			if cache.LastVersion != app.Version || cache.LastChangelog != app.WhatsNew {
				bd := map[string]interface{}{"embeds": []map[string]interface{}{map[string]interface{}{
					"author": map[string]string{
						"name":     app.Name,
						"icon_url": app.IconURL,
						"url":      "https://play.google.com/store/apps/details?id=com.discord",
					},
					"title":       "New version: **" + app.Version + "**",
					"description": strings.Replace(html.UnescapeString(app.WhatsNewHTML), "<br/>", "\n", -1),
					"footer": map[string]string{
						"text": "Updated " + app.Updated,
					},
					"color": 7506394,
				}}}
				body, _ := json.Marshal(bd)
				log("Found new version " + app.Version + ", Sending")
				http.Post(config.Webhook, "application/json", bytes.NewReader(body))
			}
		}

		if cache.LastVersion != app.Version || cache.LastChangelog != app.WhatsNew {
			cache.LastVersion = app.Version
			cache.LastChangelog = app.WhatsNew
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
