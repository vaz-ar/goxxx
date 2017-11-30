// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
//
// See LICENSE file.
package xkcd

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"log"
	"regexp"
	"testing"
)

var (
	// Mock files
	// mockFile string = "./tests_data/xkcd_1024.json" // JSON for comic 1024 (http://xkcd.com/1024/info.0.json)

	// Expected results
	expectedResult = xkcd{
		Img:   "https://imgs.xkcd.com/comics/error_code.png",
		Link:  "https://xkcd.com/1024/",
		Num:   1024,
		Title: "Error Code"}

	// IRC Events
	validEvent = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", fmt.Sprintf(" \t  !xkcd   %d   ", expectedResult.Num)}}

	validEventLastComic = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", "  !xkcd      "}}

	validEventNoResult = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", " \t  !xkcd    1000000000000000000"}}

	// Reply structs
	validReply = core.ReplyCallbackData{
		Target:  "#test_channel",
		Message: fmt.Sprintf("XKCD Comic #%d: %s => %s", expectedResult.Num, expectedResult.Title, expectedResult.Link)}

	validReplyNoResult = core.ReplyCallbackData{
		Target:  "#test_channel",
		Message: fmt.Sprintf("There is no XKCD comic #1000000000000000000")}

	reValidReplyLastComic = regexp.MustCompile(`Last XKCD Comic: (\S+\s+)+=> \S+`)
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

func Test_handleXKCDCmd(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// --- --- --- --- --- --- valid result
	var testReply core.ReplyCallbackData
	handleXKCDCmd(&validEvent, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- valid result - Last Comic
	testReply = core.ReplyCallbackData{}
	handleXKCDCmd(&validEventLastComic, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if !reValidReplyLastComic.MatchString(testReply.Message) {
		t.Errorf("Regexp %q not matching %q", reValidReplyLastComic.String(), testReply.Message)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- no result
	testReply = core.ReplyCallbackData{}
	handleXKCDCmd(&validEventNoResult, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != validReplyNoResult {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReplyNoResult)
	}
	// --- --- --- --- --- ---
}
