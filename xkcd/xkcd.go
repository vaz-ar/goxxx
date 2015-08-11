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
	URL_JSON        string = "https://xkcd.com/%d/info.0.json"
	URL_JSON_LATEST string = "https://xkcd.com/info.0.json"
)

type xkcd struct {
	Img   string `json:"img"`
	Link  string `json:"link"`
	Num   int64  `json:"num"`
	Title string `json:"title"`
}

func getComic(number int64) *xkcd {
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
		log.Println("XKCD: Not an XKCD command")
		return false
	}

	var message string
	if count < 2 {
		comic := getComic(0)
		if comic == nil {
			log.Println("XKCD: No comic return by getComic")
			return false
		}
		message = fmt.Sprintf("Last XKCD Comic: %s => %s", comic.Title, comic.Link)
	} else {
		number, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			log.Println(err)
			return false
		}

		if number < 0 || getComic(0).Num < number {
			message = fmt.Sprintf("There is no XKCD comic #%d", number)
		} else {
			comic := getComic(number)
			if comic == nil {
				log.Println("XKCD: No comic return by getComic")
				return false
			}
			message = fmt.Sprintf("XKCD Comic #%d: %s => %s", comic.Num, comic.Title, comic.Link)
		}
	}
	log.Println(message)
	callback(&core.ReplyCallbackData{Message: message, Nick: event.Nick})
	return true
}
