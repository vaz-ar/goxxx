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
	expectedResult string = "https://en.wikipedia.org/wiki/Unit_testing"
	mockFile       string = "./tests_data/duckduckgo.html" // HTML page for the search "Unit Tests"

	// Search terms
	terms          string = "Unit Tests"
	termsNoResults string = "lfdsfahlkdhfaklfa"

	// Messages
	validMessage          string = fmt.Sprintf(" \t  !dg   %s      ", terms)
	validMessageNoResults string = fmt.Sprintf("!dg %s", termsNoResults)
	invalidMessage        string = "this is not a search command"

	// IRC Events
	validEvent irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", validMessage}}

	validEventNoResults irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", validMessageNoResults}}

	invalidEvent irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", invalidMessage}}

	// Reply structs
	validReply core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("DuckDuckGo: Best result for %q => %s", terms, expectedResult)}

	validReplyNoResults core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "",
		Message: fmt.Sprintf("DuckDuckGo: No result for %q", termsNoResults)}
)

func Test_getDuckduckgoPage(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if getDuckduckgoPage(&terms) == nil {
		t.Errorf("No page returned for the search terms %q", terms)
	}
}

func Test_getDuckduckgoResult(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	content, err := ioutil.ReadFile(mockFile)
	if err != nil {
		t.Log("Failed to open the html file")
		t.FailNow()
	}

	if result := getDuckduckgoResult(content); result == nil {
		t.Error("No result returned by getDuckduckgoResult")
	} else if string(result) != expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", expectedResult, result)
	}
}

func Test_getDuckduckgoSearchResult(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if result := getDuckduckgoSearchResult(&terms); result == nil {
		t.Error("No result returned by getDuckduckgoSearchResult")
	} else if string(result) != expectedResult {
		t.Errorf("Expected result: %q, got %q instead\n", expectedResult, result)
	}
}

func Test_HandleSearchCmd_ddg(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	// Initialise search map
	Init()

	// --- --- --- --- --- --- valid result
	var testReply core.ReplyCallbackData
	HandleSearchCmd(&validEvent, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = core.ReplyCallbackData{}
	HandleSearchCmd(&validEventNoResults, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != validReplyNoResults {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReplyNoResults)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no search command
	HandleSearchCmd(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain a search command (Message: %q)\n\n", invalidMessage)
	})
	// --- --- --- --- --- ---
}
