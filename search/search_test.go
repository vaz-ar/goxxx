// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package search

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"io/ioutil"
	"testing"
)

var (
	// Search terms
	searchTerms                 = "Unit Testing"
	urbanDictionnarySearchTerms = "smh"
	searchTermsNoResults        = "lfdsfahlkdhfaklfa"

	// Mock files
	ddgMockFile              = "./tests_data/duckduckgo.html"       // HTML page for the search "Unit Testing"
	wikipediaMockFile        = "./tests_data/wikipedia.json"        // JSON for "Unit Testing"
	urbanDictionnaryMockFile = "./tests_data/urbanDictionnary.json" // JSON for "smh"

	// Expected results
	ddgExpectedResult              = "https://en.wikipedia.org/wiki/Unit_testing"
	wikipediaExpectedResult        = "https://en.wikipedia.org/wiki/Unit_Testing"
	urbanDictionnaryExpectedResult = "http://smh.urbanup.com/507685"

	// IRC Events - General
	invalidEvent = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			"this is not a search command"}}

	// IRC Events - DuckduckGo
	ddgValidEvent = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf(" \t  !dg   %s      ", searchTerms)}}

	ddgValidEventNoResults = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf("!dg %s", searchTermsNoResults)}}

	// IRC Events - Wikipedia
	wikipediaValidEvent = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf(" \t  !w   %s      ", searchTerms)}}

	wikipediaValidEventNoResults = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf("!w %s", searchTermsNoResults)}}

	// IRC Events - Urban Dictionnary
	urbanDictionnaryValidEvent = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf(" \t  !u   %s      ", urbanDictionnarySearchTerms)}}

	urbanDictionnaryValidEventNoResults = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf("!u %s", searchTermsNoResults)}}

	// Reply structs - DuckduckGo
	ddgValidReply = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("DuckDuckGo: Best result for %q => %s", searchTerms, ddgExpectedResult)}

	ddgValidReplyNoResults = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("DuckDuckGo: No result for %q", searchTermsNoResults)}

	// Reply structs - Wikipedia
	wikipediaValidReply = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Wikipedia result for %q => %s", searchTerms, wikipediaExpectedResult)}

	wikipediaValidReplyNoResults = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Wikipedia: No result for %q", searchTermsNoResults)}

	// Reply structs - Urban Dictionnary
	urbanDictionnaryValidReply = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Urban Dictionnary: Best result for %q => %s", urbanDictionnarySearchTerms, urbanDictionnaryExpectedResult)}

	urbanDictionnaryValidReplyNoResults = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Urban Dictionnary: No result for %q", searchTermsNoResults)}
)

// --- --- --- General --- --- ---
func Test_getResponseAsText(t *testing.T) {
	if getResponseAsText(fmt.Sprintf(duckduckgoURL, searchTerms)) == nil {
		t.Errorf("getResponseAsText: No data returned for the search terms %q", searchTerms)
	}
}

// --- --- --- DuckduckGo --- --- ---
func Test_getDuckduckgoResultFromHtml(t *testing.T) {
	content, err := ioutil.ReadFile(ddgMockFile)
	if err != nil {
		t.Log("Failed to open the html file")
		t.FailNow()
	}

	if result := getDuckduckgoResultFromHTML(content); result == "" {
		t.Error("No result returned by getDuckduckgoResultFromHtml")
	} else if string(result) != ddgExpectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", ddgExpectedResult, result)
	}
}

func Test_getDuckduckgoSearchResult(t *testing.T) {
	if result := getDuckduckgoSearchResult(searchTerms, ""); result == nil {
		t.Error("No result returned by getDuckduckgoSearchResult")
	} else if result[0] != ddgExpectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", ddgExpectedResult, result[0])
	}
}

func Test_HandleSearchCmd_DDG(t *testing.T) {
	// --- --- --- --- --- --- valid result
	var testReply core.ReplyCallbackData
	HandleSearchCmd(&ddgValidEvent, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != ddgValidReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, ddgValidReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = core.ReplyCallbackData{}
	HandleSearchCmd(&ddgValidEventNoResults, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != ddgValidReplyNoResults {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, ddgValidReplyNoResults)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no search command
	HandleSearchCmd(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain a search command (Message: %q)\n\n", invalidEvent.Arguments[1])
	})
	// --- --- --- --- --- ---
}

// --- --- --- Wikipedia --- --- ---
func Test_getWikipediaResultFromJson(t *testing.T) {
	content, err := ioutil.ReadFile(wikipediaMockFile)
	if err != nil {
		t.Log("Failed to open the json file")
		t.FailNow()
	}

	if result := getWikipediaResultFromJSON(content); result == nil {
		t.Error("No result returned by getWikipediaResultFromJson")
	} else if result[0] != wikipediaExpectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", wikipediaExpectedResult, result[0])
	}
}

func Test_getWikipediaSearchResult(t *testing.T) {
	if result := getWikipediaSearchResult(searchTerms, "en"); result == nil {
		t.Error("No result returned by getWikipediaSearchResult")
	} else if result[0] != wikipediaExpectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", wikipediaExpectedResult, result[0])
	}
}

func Test_HandleSearchCmd_W(t *testing.T) {
	// --- --- --- --- --- --- valid result
	var testReply []core.ReplyCallbackData
	HandleSearchCmd(&wikipediaValidEvent, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != wikipediaValidReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], wikipediaValidReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = []core.ReplyCallbackData{}
	HandleSearchCmd(&wikipediaValidEventNoResults, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != wikipediaValidReplyNoResults {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], wikipediaValidReplyNoResults)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no search command
	HandleSearchCmd(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain a search command (Message: %q)\n\n", invalidEvent.Arguments[1])
	})
	// --- --- --- --- --- ---
}

// --- --- --- Urban Dictionnary --- --- ---
func Test_getUrbanDictionnaryResultFromJson(t *testing.T) {
	content, err := ioutil.ReadFile(urbanDictionnaryMockFile)
	if err != nil {
		t.Log("Failed to open the json file")
		t.FailNow()
	}

	if result := getUrbanDictionnaryResultFromJSON(content); result == nil {
		t.Error("No result returned by getUrbanDictionnaryResultFromJSON")
	} else if result[0] != urbanDictionnaryExpectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", urbanDictionnaryExpectedResult, result[0])
	}
}

func Test_getUrbanDictionnarySearchResult(t *testing.T) {
	if result := getUrbanDictionnarySearchResult(urbanDictionnarySearchTerms, ""); result == nil {
		t.Error("No result returned by getUrbanDictionnarySearchResult")
	} else if result[0] != urbanDictionnaryExpectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", urbanDictionnaryExpectedResult, result[0])
	}
}

func Test_HandleSearchCmd_UD(t *testing.T) {
	// --- --- --- --- --- --- valid result
	var testReply []core.ReplyCallbackData
	HandleSearchCmd(&urbanDictionnaryValidEvent, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != urbanDictionnaryValidReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], urbanDictionnaryValidReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = []core.ReplyCallbackData{}
	HandleSearchCmd(&urbanDictionnaryValidEventNoResults, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != urbanDictionnaryValidReplyNoResults {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], urbanDictionnaryValidReplyNoResults)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no search command
	HandleSearchCmd(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain a search command (Message: %q)\n\n", invalidEvent.Arguments[1])
	})
	// --- --- --- --- --- ---
}
