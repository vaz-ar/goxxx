// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package xkcd

import (
	"encoding/json"
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	HELP_XKCD     string = "\t!xkcd \t\t\t\t\t\t=> Return the last XKCD comic"
	HELP_XKCD_NUM string = "\t!xkcd <comic number> \t\t=> Return the XKCD comic corresponding to the number"

	URL_SITE        string = "https://xkcd.com/%d/"
	URL_JSON        string = "http://xkcd.com/%d/info.0.json"
	URL_JSON_LATEST string = "http://xkcd.com/info.0.json"
)

type xkcd struct {
	// Alt string `json:"alt"`
	// Day  string `json:"day"`
	Img  string `json:"img"`
	Link string `json:"link"`
	// Month      string `json:"month"`
	// News       string `json:"news"`
	Num int `json:"num"`
	// SafeTitle  string `json:"safe_title"`
	Title string `json:"title"`
	// Transcript string `json:"transcript"`
	// Year       string `json:"year"`
}

// func init() {
// }

func getComic(number int) *xkcd {
	var url string
	if number <= 0 { // Get latest comic
		url = URL_JSON_LATEST
	} else { // Get the comic corresponding to "number"
		url = fmt.Sprintf(URL_JSON, number)
	}

	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer response.Body.Close()

	jsonDataFromHttp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	response.Body.Close()

	result := new(xkcd)
	err = json.Unmarshal(jsonDataFromHttp, result)
	if err != nil {
		log.Println(err)
		return nil
	}

	result.Link = fmt.Sprintf(URL_SITE, result.Num)

	return result
}

func HandleXKCDCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	if callback == nil {
		log.Println("Callback nil for the HandleXKCDCmd function")
		return false
	}

	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1]  => Comic #

	count := len(fields)
	if count == 0 || fields[0] != "!xkcd" {
		return false
	}

	var message string
	if count < 2 {
		comic := getComic(0)
		if comic == nil {
			return false
		}
		message = fmt.Sprintf("Last XKCD Comic: %s => %s", comic.Title, comic.Link)
	} else {
		number, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Println(err)
			return false
		}
		if number < 0 || getComic(0).Num < number {
			message = fmt.Sprintf("there is no XKCD comic #%d", number)
		} else {
			comic := getComic(number)
			if comic == nil {
				return false
			}
			message = fmt.Sprintf("XKCD Comic #%d: %s => %s", comic.Num, comic.Title, comic.Link)
		}
	}
	callback(&core.ReplyCallbackData{Message: message, Nick: event.Nick})
	return true
}
