// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
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

const (
	// Duckduckgo URL format string
	duckduckgoURL = "https://duckduckgo.com/html/?q=%s"
	// Wikipedia URL format string
	wikipediaURL = "https://%s.wikipedia.org/w/api.php?format=json&action=query&prop=extracts|info&exintro=&explaintext=&inprop=url&titles=%s"
	// Urban Dictionnary URL format string
	urbanDictionnaryURL = "http://api.urbandictionary.com/v0/define?term=%s"
)

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

// GetDuckduckGoCmd returns a Command structure for the duckduckGo command
func GetDuckduckGoCmd() *core.Command {
	return &core.Command{
		Module:      "search",
		HelpMessage: "!d/!dg/!ddg <terms to search> => Search on DuckduckGo",
		Triggers:    []string{"!d", "!dg", "!ddg"},
		Handler:     handleDuckduckGoCmd}
}

// GetWikipediaCmd returns a Command structure for the wikipedia command
func GetWikipediaCmd() *core.Command {
	return &core.Command{
		Module:      "search",
		HelpMessage: "!w/!wiki <terms to search> => Search on Wikipedia EN",
		Triggers:    []string{"!w", "!wiki"},
		Handler:     handleWikipediaCmd}
}

// GetWikipediaFRCmd returns a Command structure for the wikipedia command
func GetWikipediaFRCmd() *core.Command {
	return &core.Command{
		Module:      "search",
		HelpMessage: "!wf/!wfr <terms to search> => Search on Wikipedia FR",
		Triggers:    []string{"!wf", "!wfr"},
		Handler:     handleWikipediaCmd}
}

// GetUrbanDictionnaryCmd returns a Command structure for the urban dictionnary command
func GetUrbanDictionnaryCmd() *core.Command {
	return &core.Command{
		Module:      "search",
		HelpMessage: "!u/!ud <terms to search> => Search on Urban Dictionnary",
		Triggers:    []string{"!u", "!ud"},
		Handler:     handleUrbanDictionnaryCmd}
}

// handleDuckduckGoCmd handles the duckduckGo search command
func handleDuckduckGoCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => terms to search for
	if len(fields) < 2 {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Search usage: %s \"terms to search for\"", fields[0]),
			Target:  core.GetTargetFromEvent(event)})
		return false
	}
	message := strings.Join(fields[1:], " ")
	results := getDuckduckgoSearchResult(message)
	if results == nil {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("DuckDuckGo: No result for \"%s\"", message),
			Target:  core.GetTargetFromEvent(event)})
		return true
	}
	for index, item := range results {
		if index == 0 {
			// First part of the result is sent to everyone, no nick is sent
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("DuckDuckGo: Best result for \"%s\" => %s", message, item),
				Target:  core.GetTargetFromEvent(event)})
		} else {
			// Second and following parts of the result are sent directly to the user
			callback(&core.ReplyCallbackData{Target: event.Nick, Message: item})
		}
	}
	return true
}

// handleUrbanDictionnaryCmd handles the urban dictionnary search command
func handleUrbanDictionnaryCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => terms to search for
	if len(fields) < 2 {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Search usage: %s \"terms to search for\"", fields[0]),
			Target:  core.GetTargetFromEvent(event)})
		return false
	}
	message := strings.Join(fields[1:], " ")
	results := getUrbanDictionnarySearchResult(message)

	if results == nil {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Urban Dictionnary: No result for \"%s\"", message),
			Target:  core.GetTargetFromEvent(event)})
		return true
	}
	for index, item := range results {
		if index == 0 {
			// First part of the result is sent to everyone, no nick is sent
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("Urban Dictionnary: Best result for \"%s\" => %s", message, item),
				Target:  core.GetTargetFromEvent(event)})
		} else {
			// Second and following parts of the result are sent directly to the user
			callback(&core.ReplyCallbackData{
				Target:  event.Nick,
				Message: fmt.Sprintf("Definition: %s", item)})
		}
	}
	return true
}

// handleWikipediaCmd handles the wikipedia command
func handleWikipediaCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => terms to search for
	if len(fields) < 2 {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Search usage: %s \"terms to search for\"", fields[0]),
			Target:  core.GetTargetFromEvent(event)})
		return false
	}
	message := strings.Join(fields[1:], " ")
	results := getWikipediaSearchResult(message, "en")
	if results == nil {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Wikipedia: No result for \"%s\"", message),
			Target:  core.GetTargetFromEvent(event)})
		return true
	}
	for index, item := range results {
		if index == 0 {
			// First part of the result is sent to everyone, no nick is sent
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("Wikipedia result for \"%s\" => %s", message, item),
				Target:  core.GetTargetFromEvent(event)})
		} else {
			// Second and following parts of the result are sent directly to the user
			callback(&core.ReplyCallbackData{Target: event.Nick, Message: item})
		}
	}
	return true
}

// handleWikipediaFRCmd handles the wikipedia FR command
func handleWikipediaFRCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => terms to search for
	if len(fields) < 2 {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Search usage: %s \"terms to search for\"", fields[0]),
			Target:  core.GetTargetFromEvent(event)})
		return false
	}
	message := strings.Join(fields[1:], " ")
	results := getWikipediaSearchResult(message, "fr")
	if results == nil {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Wikipedia: No result for \"%s\"", message),
			Target:  core.GetTargetFromEvent(event)})
		return true
	}
	for index, item := range results {
		if index == 0 {
			// First part of the result is sent to everyone, no nick is sent
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("Wikipedia result for \"%s\" => %s", message, item),
				Target:  core.GetTargetFromEvent(event)})
		} else {
			// Second and following parts of the result are sent directly to the user
			callback(&core.ReplyCallbackData{Target: event.Nick, Message: item})
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

func getDuckduckgoSearchResult(searchTerms string) []string {
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

func getUrbanDictionnarySearchResult(searchTerms string) []string {
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
