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
	URL_DUCKDUCKGO       string = "https://duckduckgo.com/html/?q=%s"
	URL_WIKIPEDIA        string = "https://%s.wikipedia.org/w/api.php?format=json&action=query&prop=extracts|info&exintro=&explaintext=&inprop=url&titles=%s"
	URL_URBANDICTIONNARY string = "http://api.urbandictionary.com/v0/define?term=%s"

	HELP_DUCKDUCKGO       string = "\t!d/!dg/!ddg <terms to search> \t=> Search on DuckduckGo"
	HELP_WIKIPEDIA        string = "\t!w <terms to search> \t\t\t=> Search on Wikipedia EN"
	HELP_WIKIPEDIA_FR     string = "\t!wf/!wfr <terms to search> \t=> Search on Wikipedia FR"
	HELP_URBANDICTIONNARY string = "\t!u/!ud <terms to search> \t\t=> Search on Urban Dictionnary"
)

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

//
type searchData struct {
	getUrl         func(string, string) []string
	extraParameter string
	// First string is for the first result (URL), second string for the second result (Details/Definition/...)
	text_result    [2]string
	text_no_result string
}

// --- --- --- Global variable --- --- ---
var searchMap = make(map[string]*searchData)

func init() {
	ddg := &searchData{
		getUrl:         getDuckduckgoSearchResult,
		text_result:    [2]string{"DuckDuckGo: Best result for %q => %s"},
		text_no_result: "DuckDuckGo: No result for %q"}

	searchMap["!d"] = ddg
	searchMap["!dg"] = ddg
	searchMap["!ddg"] = ddg

	searchMap["!w"] = &searchData{
		getUrl:         getWikipediaSearchResult,
		extraParameter: "en",
		text_result:    [2]string{"Wikipedia result for %q => %s"},
		text_no_result: "Wikipedia: No result for %q"}

	wfr := &searchData{
		getUrl:         getWikipediaSearchResult,
		extraParameter: "fr",
		text_result:    [2]string{"Wikipedia result for %q => %s"},
		text_no_result: "Wikipedia: No result for %q"}

	searchMap["!wf"] = wfr
	searchMap["!wfr"] = wfr

	ud := &searchData{
		getUrl:         getUrbanDictionnarySearchResult,
		text_result:    [2]string{"Urban Dictionnary: Best result for %q => %s", "Definition: %s"},
		text_no_result: "Urban Dictionnary: No result for %q"}

	searchMap["!u"] = ud
	searchMap["!ud"] = ud
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

	results := search.getUrl(message, search.extraParameter)
	if results == nil {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf(search.text_no_result, message)})
		return true
	}
	for index, item := range results {
		switch index {
		// First part of the result is sent to everyone, no nick is sent
		case 0:
			if search.text_result[index] != "" {
				callback(&core.ReplyCallbackData{Message: fmt.Sprintf(search.text_result[index], message, item)})
			} else {
				callback(&core.ReplyCallbackData{Message: item})
			}
		// Second and following parts of the result are sent directly to the user
		case 1:
			if search.text_result[index] != "" {
				callback(&core.ReplyCallbackData{Nick: event.Nick, Message: fmt.Sprintf(search.text_result[index], item)})
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
func getDuckduckgoResultFromHtml(page []byte) string {
	re := regexp.MustCompile(`<a rel="nofollow" href="(.[^"]*)">`)
	if result := re.FindSubmatch(page); result != nil && len(result) == 2 {
		return string(result[1])
	}
	return ""
}

func getDuckduckgoSearchResult(searchTerms string, extraParameter string) []string {
	HtmlPageFromHttp := getResponseAsText(fmt.Sprintf(URL_DUCKDUCKGO, searchTerms))
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

func getWikipediaSearchResult(searchTerms string, extraParameter string) []string {
	searchTerms = strings.Replace(strings.Title(searchTerms), " ", "%20", -1)
	jsonDataFromHttp := getResponseAsText(fmt.Sprintf(URL_WIKIPEDIA, extraParameter, searchTerms))
	if jsonDataFromHttp == nil {
		return nil
	}
	result := getWikipediaResultFromJson(jsonDataFromHttp)
	if result == nil {
		return nil
	}
	return result
}

// --- Urban Dictionnary
func getUrbanDictionnaryResultFromJson(jsonDataFromHttp []byte) []string {
	var result urbanDictionnary
	err := json.Unmarshal(jsonDataFromHttp, &result)
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
	jsonDataFromHttp := getResponseAsText(fmt.Sprintf(URL_URBANDICTIONNARY, searchTerms))
	if jsonDataFromHttp == nil {
		return nil
	}
	result := getUrbanDictionnaryResultFromJson(jsonDataFromHttp)
	if result == nil {
		return nil
	}
	return result
}
