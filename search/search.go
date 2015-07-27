// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package search

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type search_struct struct {
	getUrl         func(*string) []byte
	text_result    string
	text_no_result string
}

var search_map = make(map[string]search_struct)

func Init() {
	search_map["!dg"] = search_struct{
		getUrlDuckduckgo,
		"DuckDuckGo: Best result for %q => %s",
		"DuckDuckGo: No result for %q"}
}

func HandleSearchCmd(event *irc.Event, callback func(string)) bool {
	message := strings.TrimSpace(event.Message())
	if message == "" {
		return false
	}

	cmd := strings.Split(message, " ")[0]
	search, present := search_map[cmd]
	if !present {
		return false
	}

	message = strings.TrimSpace(strings.TrimPrefix(message, cmd))
	if message == "" {
		return false
	}

	result := search.getUrl(&message)

	if callback != nil {
		if result != nil {
			callback(fmt.Sprintf(search.text_result, message, result))
		} else {
			callback(fmt.Sprintf(search.text_no_result, message))
		}
	}

	return true
}

func getUrlDuckduckgo(message *string) []byte {
	response, err := http.Get(fmt.Sprintf("https://duckduckgo.com/html/?q=%s", *message))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer response.Body.Close()

	text, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	response.Body.Close()

	re := regexp.MustCompile(`<a rel="nofollow" href="(.[^"]*)">`)
	if result := re.FindSubmatch(text); result != nil && len(result) == 2 {
		return result[1]
	} else {
		return nil
	}
}
