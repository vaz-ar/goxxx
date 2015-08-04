// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package search

import (
	"encoding/json"
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

//  --- --- --- Constants --- --- ---
const (
	URL_DUCKDUCKGO string = "https://duckduckgo.com/html/?q=%s"
	URL_WIKIPEDIA  string = "https://en.wikipedia.org/w/api.php?format=json&action=query&prop=extracts|info&exintro=&explaintext=&inprop=url&titles=%s"
)

// --- --- --- Types  --- --- ---
type wikipedia struct {
	Query struct {
		Pages map[string]struct {
			Extract string `json:"extract"`
			Fullurl string `json:"fullurl"`
			Title   string `json:"title"`
		} `json:"pages"`
	} `json:"query"`
}

type searchData struct {
	getUrl         func(string) []string
	text_result    [2]string
	text_no_result string
}

// --- --- --- Global variable --- --- ---
var searchMap = make(map[string]searchData)

func Init() {
	searchMap["!dg"] = searchData{
		getDuckduckgoSearchResult,
		[2]string{"DuckDuckGo: Best result for %q => %s"},
		"DuckDuckGo: No result for %q"}
	searchMap["!w"] = searchData{
		getWikipediaSearchResult,
		[2]string{"Wikipedia result for %q => %s"},
		"Wikipedia: No result for %q"}
}

func HandleSearchCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
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

	search, present := searchMap[fields[0]]
	if !present {
		return false
	}

	message := strings.Join(fields[1:], " ")
	if message == "" {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("Search usage: %s \"terms to search for\"", fields[0])})
		return false
	}

	results := search.getUrl(message)
	if results == nil {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf(search.text_no_result, message)})
		return true
	}
	for i, item := range results {
		if i == 0 && search.text_result[i] != "" {
			callback(&core.ReplyCallbackData{Message: fmt.Sprintf(search.text_result[i], message, item)})
		} else if i == 1 && search.text_result[i] != "" {
			callback(&core.ReplyCallbackData{Message: fmt.Sprintf(search.text_result[i], item)})
		} else {
			callback(&core.ReplyCallbackData{Message: item})
		}
	}
	return true
}

// --- --- --- HTTP Functions --- --- ---
func getResponseAsText(url string, searchTerms *string) []byte {
	response, err := http.Get(fmt.Sprintf(url, *searchTerms))
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
	return text
}

// --- --- --- Search provider functions --- --- ---

// --- Duckduckgo
func getDuckduckgoResultFromHtml(page []byte) string {
	re := regexp.MustCompile(`<a rel="nofollow" href="(.[^"]*)">`)
	if result := re.FindSubmatch(page); result != nil && len(result) == 2 {
		return string(result[1])
	}
	return ""
}

func getDuckduckgoSearchResult(searchTerms string) []string {
	HtmlPageFromHttp := getResponseAsText(URL_DUCKDUCKGO, &searchTerms)
	if HtmlPageFromHttp == nil {
		return nil
	}
	result := getDuckduckgoResultFromHtml(HtmlPageFromHttp)
	if result == "" {
		return nil
	}
	return []string{result}
}

// --- Wikipedia
func getWikipediaResultFromJson(jsonDataFromHttp []byte) []string {
	var result wikipedia
	err := json.Unmarshal(jsonDataFromHttp, &result)
	if err != nil {
		log.Println(err)
		return nil
	}

	var returnValues []string
	for key, value := range result.Query.Pages {
		if key != "-1" {
			returnValues = append(returnValues, value.Fullurl)
			returnValues = append(returnValues, strings.Split(value.Extract, ". ")...)
		}
	}
	return returnValues
}

func getWikipediaSearchResult(searchTerms string) []string {
	searchTerms = strings.Replace(strings.Title(searchTerms), " ", "%20", -1)
	jsonDataFromHttp := getResponseAsText(URL_WIKIPEDIA, &searchTerms)
	if jsonDataFromHttp == nil {
		return nil
	}
	result := getWikipediaResultFromJson(jsonDataFromHttp)
	if result == nil {
		return nil
	}
	return result
}
