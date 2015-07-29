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

type searchData struct {
	getUrl         func(*string) []byte
	text_result    string
	text_no_result string
}

var search_map = make(map[string]searchData)

func Init() {
	search_map["!dg"] = searchData{
		getUrlDuckduckgo,
		"DuckDuckGo: Best result for %q => %s",
		"DuckDuckGo: No result for %q"}
}

func HandleSearchCmd(event *irc.Event, callback func(string)) bool {
	if callback == nil {
		log.Println("Callback nil for the HandleSearchCmd function")
		return false
	}

	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => terms to search for
	if len(fields) == 0 {
		return false
	}

	search, present := search_map[fields[0]]
	if !present {
		return false
	}

	message := strings.Join(fields[1:], " ")
	if message == "" {
		callback(fmt.Sprintf("Search usage: %s \"terms to search for\"", fields[0]))
		return false
	}

	if result := search.getUrl(&message); result != nil {
		callback(fmt.Sprintf(search.text_result, message, result))
	} else {
		callback(fmt.Sprintf(search.text_no_result, message))
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
