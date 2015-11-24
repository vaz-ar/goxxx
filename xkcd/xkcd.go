// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Package xkcd retrieves XKCD Comics
package xkcd

import (
	"encoding/json"
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	urlWebsite    string = "https://xkcd.com/%d/"            // Website URL format string
	urlJSON       string = "https://xkcd.com/%d/info.0.json" // JSON URL format string
	urlLatestJSON string = "https://xkcd.com/info.0.json"    // JSON URL for the current comic
)

type xkcd struct {
	Img   string `json:"img"`
	Link  string `json:"link"`
	Num   int64  `json:"num"`
	Title string `json:"title"`
}

// GetCommand returns a Command structure for the XKCD command
func GetCommand() *core.Command {
	return &core.Command{
		Module:      "xkcd",
		HelpMessage: "\t!xkcd [<comic number>] \t\t=> Return the XKCD comic corresponding to the number. If number is not specified, returns the last comic.",
		Triggers:    []string{"!xkcd"},
		Handler:     handleXKCDCmd}
}

// If number is superior to 0 attempt to get informations on the corresponding comic, else return the inforamtions for the current comic.
// In case of error return nil
func getComic(number int64) *xkcd {
	var url string
	if number <= 0 {
		// Get latest comic
		url = urlLatestJSON
	} else {
		// Get the comic corresponding to "number"
		url = fmt.Sprintf(urlJSON, number)
	}

	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer response.Body.Close()

	jsonDataFromHTTP, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	response.Body.Close()

	result := new(xkcd)
	err = json.Unmarshal(jsonDataFromHTTP, result)
	if err != nil {
		log.Println(err)
		return nil
	}
	// Add the full website link to the structure
	result.Link = fmt.Sprintf(urlWebsite, result.Num)
	return result
}

// handleXKCDCmd Handles XKCD commands
func handleXKCDCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
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
	callback(&core.ReplyCallbackData{Message: message, Target: event.Nick})
	return true
}
