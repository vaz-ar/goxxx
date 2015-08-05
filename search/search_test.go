// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package search

import (
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"testing"
)

var (
	// Search terms
	searchTerms          string = "Unit Testing"
	UD_searchTerms       string = "smh"
	searchTermsNoResults string = "lfdsfahlkdhfaklfa"

	// Mock files
	DDG_mockFile string = "./tests_data/duckduckgo.html"       // HTML page for the search "Unit Testing"
	W_mockFile   string = "./tests_data/wikipedia.json"        // JSON for "Unit Testing"
	UD_mockFile  string = "./tests_data/urbanDictionnary.json" // JSON for "smh"

	// Expected results
	DDG_expectedResult string = "https://en.wikipedia.org/wiki/Unit_testing"
	W_expectedResult   string = "https://en.wikipedia.org/wiki/Unit_Testing"
	UD_expectedResult  string = "http://smh.urbanup.com/507685"

	// IRC Events - General
	invalidEvent irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			"this is not a search command"}}

	// IRC Events - DuckduckGo
	DDG_validEvent irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf(" \t  !dg   %s      ", searchTerms)}}

	DDG_validEventNoResults irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf("!dg %s", searchTermsNoResults)}}

	// IRC Events - Wikipedia
	W_validEvent irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf(" \t  !w   %s      ", searchTerms)}}

	W_validEventNoResults irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf("!w %s", searchTermsNoResults)}}

	// IRC Events - Urban Dictionnary
	UD_validEvent irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf(" \t  !u   %s      ", UD_searchTerms)}}

	UD_validEventNoResults irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			fmt.Sprintf("!u %s", searchTermsNoResults)}}

	// Reply structs - DuckduckGo
	DDG_validReply core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("DuckDuckGo: Best result for %q => %s", searchTerms, DDG_expectedResult)}

	DDG_validReplyNoResults core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("DuckDuckGo: No result for %q", searchTermsNoResults)}

	// Reply structs - Wikipedia
	W_validReply core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Wikipedia result for %q => %s", searchTerms, W_expectedResult)}

	W_validReplyNoResults core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Wikipedia: No result for %q", searchTermsNoResults)}

	// Reply structs - Urban Dictionnary
	UD_validReply core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Urban Dictionnary: Best result for %q => %s", UD_searchTerms, UD_expectedResult)}

	UD_validReplyNoResults core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("Urban Dictionnary: No result for %q", searchTermsNoResults)}
)

// --- --- --- General --- --- ---
func Test_getResponseAsText(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if getResponseAsText(fmt.Sprintf(URL_DUCKDUCKGO, searchTerms)) == nil {
		t.Errorf("getResponseAsText: No data returned for the search terms %q", searchTerms)
	}
}

// --- --- --- DuckduckGo --- --- ---
func Test_getDuckduckgoResultFromHtml(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	content, err := ioutil.ReadFile(DDG_mockFile)
	if err != nil {
		t.Log("Failed to open the html file")
		t.FailNow()
	}

	if result := getDuckduckgoResultFromHtml(content); result == "" {
		t.Error("No result returned by getDuckduckgoResultFromHtml")
	} else if string(result) != DDG_expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", DDG_expectedResult, result)
	}
}

func Test_getDuckduckgoSearchResult(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if result := getDuckduckgoSearchResult(searchTerms, ""); result == nil {
		t.Error("No result returned by getDuckduckgoSearchResult")
	} else if result[0] != DDG_expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", DDG_expectedResult, result[0])
	}
}

func Test_HandleSearchCmd_DDG(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Initialise search map
	Init()

	// --- --- --- --- --- --- valid result
	var testReply core.ReplyCallbackData
	HandleSearchCmd(&DDG_validEvent, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != DDG_validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, DDG_validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = core.ReplyCallbackData{}
	HandleSearchCmd(&DDG_validEventNoResults, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != DDG_validReplyNoResults {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, DDG_validReplyNoResults)
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
	log.SetOutput(ioutil.Discard)

	content, err := ioutil.ReadFile(W_mockFile)
	if err != nil {
		t.Log("Failed to open the json file")
		t.FailNow()
	}

	if result := getWikipediaResultFromJson(content); result == nil {
		t.Error("No result returned by getWikipediaResultFromJson")
	} else if result[0] != W_expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", W_expectedResult, result[0])
	}
}

func Test_getWikipediaSearchResult(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if result := getWikipediaSearchResult(searchTerms, "en"); result == nil {
		t.Error("No result returned by getWikipediaSearchResult")
	} else if result[0] != W_expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", W_expectedResult, result[0])
	}
}

func Test_HandleSearchCmd_W(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Initialise search map
	Init()

	// --- --- --- --- --- --- valid result
	var testReply []core.ReplyCallbackData
	HandleSearchCmd(&W_validEvent, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != W_validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], W_validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = []core.ReplyCallbackData{}
	HandleSearchCmd(&W_validEventNoResults, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != W_validReplyNoResults {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], W_validReplyNoResults)
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
	log.SetOutput(ioutil.Discard)

	content, err := ioutil.ReadFile(UD_mockFile)
	if err != nil {
		t.Log("Failed to open the json file")
		t.FailNow()
	}

	if result := getUrbanDictionnaryResultFromJson(content); result == nil {
		t.Error("No result returned by getUrbanDictionnaryResultFromJson")
	} else if result[0] != UD_expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", UD_expectedResult, result[0])
	}
}

func Test_getUrbanDictionnarySearchResult(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if result := getUrbanDictionnarySearchResult(UD_searchTerms, ""); result == nil {
		t.Error("No result returned by getUrbanDictionnarySearchResult")
	} else if result[0] != UD_expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", UD_expectedResult, result[0])
	}
}

func Test_HandleSearchCmd_UD(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Initialise search map
	Init()

	// --- --- --- --- --- --- valid result
	var testReply []core.ReplyCallbackData
	HandleSearchCmd(&UD_validEvent, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != UD_validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], UD_validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = []core.ReplyCallbackData{}
	HandleSearchCmd(&UD_validEventNoResults, func(data *core.ReplyCallbackData) {
		testReply = append(testReply, *data)
	})
	if testReply[0] != UD_validReplyNoResults {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply[0], UD_validReplyNoResults)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no search command
	HandleSearchCmd(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain a search command (Message: %q)\n\n", invalidEvent.Arguments[1])
	})
	// --- --- --- --- --- ---
}
