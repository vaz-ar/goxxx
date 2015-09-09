// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

/*
Package search allow to do web searches
Current sources:
	- DuckduckGo
	- Urban Dictionnary
	- Wikipedia EN
	- Wikipedia FR
*/
package search

import (
	"encoding/json"
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

//  --- --- --- Constants --- --- ---
const (
	duckduckgoURL       = "https://duckduckgo.com/html/?q=%s"                                                                                         // Duckduckgo URL format string
	wikipediaURL        = "https://%s.wikipedia.org/w/api.php?format=json&action=query&prop=extracts|info&exintro=&explaintext=&inprop=url&titles=%s" // Wikipedia URL format string
	urbanDictionnaryURL = "http://api.urbandictionary.com/v0/define?term=%s"                                                                          // Urban Dictionnary URL format string

	HelpDuckduckgo       = "\t!d/!dg/!ddg <terms to search> \t=> Search on DuckduckGo"     // Help message for the Duckduckgo commands
	HelpWikipedia        = "\t!w <terms to search> \t\t\t=> Search on Wikipedia EN"        // Help message for the Wikipedia EN commands
	HelpWikipediaFr      = "\t!wf/!wfr <terms to search> \t=> Search on Wikipedia FR"      // Help message for the Wikipedia FR commands
	HelpUrbanDictionnary = "\t!u/!ud <terms to search> \t\t=> Search on Urban Dictionnary" // Help message for the Urban Dictionnary commands
)

// --- --- --- Global variable --- --- ---
var searchMap = make(map[string]*searchData)

// --- --- --- Types --- --- ---
// Wikipedia JSON struct
type wikipedia struct {
	Query struct {
		Pages map[string]struct {
			Extract string `json:"extract"`
			Fullurl string `json:"fullurl"`
			Title   string `json:"title"`
		} `json:"pages"`
	} `json:"query"`
}

// Urban Dictionnary JSON struct
type urbanDictionnary struct {
	List []struct {
		Definition string `json:"definition"`
		Example    string `json:"example"`
		Permalink  string `json:"permalink"`
	} `json:"list"`
}

// General data structure to store search informations
type searchData struct {
	getURL         func(string, string) []string
	extraParameter string
	// First string is for the first result (URL), second string for the second result (Details/Definition/...)
	textResult   [2]string
	textNoResult string
}

// --- --- --- Functions --- --- ---

func init() {
	ddg := &searchData{
		getURL:       getDuckduckgoSearchResult,
		textResult:   [2]string{"DuckDuckGo: Best result for %q => %s"},
		textNoResult: "DuckDuckGo: No result for %q"}

	searchMap["!d"] = ddg
	searchMap["!dg"] = ddg
	searchMap["!ddg"] = ddg

	searchMap["!w"] = &searchData{
		getURL:         getWikipediaSearchResult,
		extraParameter: "en",
		textResult:     [2]string{"Wikipedia result for %q => %s"},
		textNoResult:   "Wikipedia: No result for %q"}

	wfr := &searchData{
		getURL:         getWikipediaSearchResult,
		extraParameter: "fr",
		textResult:     [2]string{"Wikipedia result for %q => %s"},
		textNoResult:   "Wikipedia: No result for %q"}

	searchMap["!wf"] = wfr
	searchMap["!wfr"] = wfr

	ud := &searchData{
		getURL:       getUrbanDictionnarySearchResult,
		textResult:   [2]string{"Urban Dictionnary: Best result for %q => %s", "Definition: %s"},
		textNoResult: "Urban Dictionnary: No result for %q"}

	searchMap["!u"] = ud
	searchMap["!ud"] = ud
}

// HandleSearchCmd handles all the search commands
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

	results := search.getURL(message, search.extraParameter)
	if results == nil {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf(search.textNoResult, message)})
		return true
	}
	for index, item := range results {
		switch index {
		// First part of the result is sent to everyone, no nick is sent
		case 0:
			if search.textResult[index] != "" {
				callback(&core.ReplyCallbackData{Message: fmt.Sprintf(search.textResult[index], message, item)})
			} else {
				callback(&core.ReplyCallbackData{Message: item})
			}
		// Second and following parts of the result are sent directly to the user
		case 1:
			if search.textResult[index] != "" {
				callback(&core.ReplyCallbackData{Nick: event.Nick, Message: fmt.Sprintf(search.textResult[index], item)})
			} else {
				callback(&core.ReplyCallbackData{Nick: event.Nick, Message: item})
			}
		default:
			callback(&core.ReplyCallbackData{Nick: event.Nick, Message: item})
		}
	}
	return true
}

// --- --- --- HTTP Functions --- --- ---

// Function to get text content from an url
func getResponseAsText(url string) []byte {
	response, err := http.Get(url)
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

func getDuckduckgoResultFromHTML(page []byte) string {
	re := regexp.MustCompile(`<a rel="nofollow" href="(.[^"]*)">`)
	if result := re.FindSubmatch(page); result != nil && len(result) == 2 {
		return string(result[1])
	}
	return ""
}

func getDuckduckgoSearchResult(searchTerms string, extraParameter string) []string {
	webPage := getResponseAsText(fmt.Sprintf(duckduckgoURL, searchTerms))
	if webPage == nil {
		return nil
	}
	result := getDuckduckgoResultFromHTML(webPage)
	if result == "" {
		return nil
	}
	return []string{result}
}

// --- Wikipedia
func getWikipediaResultFromJSON(jsonDataFromHTTP []byte) []string {
	var result wikipedia
	err := json.Unmarshal(jsonDataFromHTTP, &result)
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

func getWikipediaSearchResult(searchTerms string, extraParameter string) []string {
	searchTerms = strings.Replace(strings.Title(searchTerms), " ", "%20", -1)
	jsonDataFromHTTP := getResponseAsText(fmt.Sprintf(wikipediaURL, extraParameter, searchTerms))
	if jsonDataFromHTTP == nil {
		return nil
	}
	result := getWikipediaResultFromJSON(jsonDataFromHTTP)
	if result == nil {
		return nil
	}
	return result
}

// --- Urban Dictionnary
func getUrbanDictionnaryResultFromJSON(jsonDataFromHTTP []byte) []string {
	var result urbanDictionnary
	err := json.Unmarshal(jsonDataFromHTTP, &result)
	if err != nil {
		log.Println(err)
		return nil
	}
	if len(result.List) == 0 {
		return nil
	}
	returnValues := []string{result.List[0].Permalink}
	returnValues = append(returnValues, strings.Split(result.List[0].Definition, ". ")...)

	return returnValues
}

func getUrbanDictionnarySearchResult(searchTerms string, extraParameter string) []string {
	searchTerms = strings.Replace(strings.Title(searchTerms), " ", "%20", -1)
	jsonDataFromHTTP := getResponseAsText(fmt.Sprintf(urbanDictionnaryURL, searchTerms))
	if jsonDataFromHTTP == nil {
		return nil
	}
	result := getUrbanDictionnaryResultFromJSON(jsonDataFromHTTP)
	if result == nil {
		return nil
	}
	return result
}
