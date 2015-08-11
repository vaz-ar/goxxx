// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package xkcd

import (
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/thoj/go-ircevent"
	"log"
	"regexp"
	"testing"
)

var (
	// Mock files
	// mockFile string = "./tests_data/xkcd_1024.json" // JSON for comic 1024 (http://xkcd.com/1024/info.0.json)

	// Expected results
	expectedResult xkcd = xkcd{
		Img:   "http://imgs.xkcd.com/comics/error_code.png",
		Link:  "https://xkcd.com/1024/",
		Num:   1024,
		Title: "Error Code"}

	// IRC Events
	invalidEvent irc.Event = irc.Event{
		Nick: "Sender",
		Arguments: []string{
			"#test_channel",
			"this is not a !xkcd command"}}

	validEvent irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", fmt.Sprintf(" \t  !xkcd   %d   ", expectedResult.Num)}}

	validEventLastComic irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", "  !xkcd      "}}

	validEventNoResult irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", " \t  !xkcd    1000000000000000000"}}

	// Reply structs
	validReply core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "Sender",
		Message: fmt.Sprintf("XKCD Comic #%d: %s => %s", expectedResult.Num, expectedResult.Title, expectedResult.Link)}

	validReplyNoResult core.ReplyCallbackData = core.ReplyCallbackData{
		Nick:    "Sender",
		Message: fmt.Sprintf("There is no XKCD comic #1000000000000000000")}

	re_validReplyLastComic *regexp.Regexp = regexp.MustCompile(`Last XKCD Comic: (\w+\s+)+=> \S+`)
)

// --- --- --- General --- --- ---
func Test_getComic(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	comic := getComic(expectedResult.Num)
	if comic == nil {
		t.Errorf("getComic: No data returned")
	} else if *comic != expectedResult {
		t.Errorf("getComic:Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", comic, expectedResult)
	}
}

func Test_HandleXKCDCmd(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// --- --- --- --- --- --- valid result
	var testReply core.ReplyCallbackData
	HandleXKCDCmd(&validEvent, func(data *core.ReplyCallbackData) {
		fmt.Sprintf("\t%#v\n", data)
		testReply = *data
	})
	if testReply != validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- valid result - Last Comic
	testReply = core.ReplyCallbackData{}
	HandleXKCDCmd(&validEventLastComic, func(data *core.ReplyCallbackData) {
		fmt.Sprintf("\t%#v\n", data)
		testReply = *data
	})
	if !re_validReplyLastComic.MatchString(testReply.Message) {
		t.Errorf("Regexp %q not matching %q", re_validReplyLastComic.String(), testReply.Message)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = core.ReplyCallbackData{}
	HandleXKCDCmd(&validEventNoResult, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != validReplyNoResult {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReplyNoResult)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no search command
	HandleXKCDCmd(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain a search command (Message: %q)\n\n", invalidEvent.Arguments[1])
	})
	// --- --- --- --- --- ---
}
