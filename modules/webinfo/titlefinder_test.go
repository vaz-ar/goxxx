// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
//
// See LICENSE file.
package webinfo

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"github.com/vaz-ar/goxxx/database"
	"golang.org/x/net/html"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

var (
	htmlsWithTitle = []string{
		"./tests_data/page_with_title.html",
		"./tests_data/page_with_title_containing_spaces.html",
		"./tests_data/page_with_multiline_title.html"}
	htmlWithoutTitle = "./tests_data/page_without_title.html"
	expectedTitle    = "The Ultimate Unicorn"
	expectedNick     = "Sender"

	messagesWithUrls = []string{
		"Oh look at this link matias.ma/nsfw/ (It is not NSFW)",
		"https://golang.org/doc/effective_go.html", // -- 1
		"Check this one https://diasporafoundation.org/, and this other one cozy.io/en/, and this last one: http://framasoft.net/"}
	expectedUrls = [][]string{
		{"http://matias.ma/nsfw/"},
		{"https://golang.org/doc/effective_go.html"},
		{"https://diasporafoundation.org/", "http://cozy.io/en/", "http://framasoft.net/"}}

	messageWithoutURL = "This is just.a.message without/any URL in.it"

	validEvent = irc.Event{
		Nick:      expectedNick,
		Arguments: []string{"#test_channel", messagesWithUrls[1]}} // -- 1

	invalidEvent = irc.Event{
		Nick:      expectedNick,
		Arguments: []string{"#test_channel", messageWithoutURL}}

	validReply = core.ReplyCallbackData{Target: "#test_channel", Message: "Effective Go - The Go Programming Language"} // -- 1

	re = regexp.MustCompile(fmt.Sprintf(`^Link already posted by %s \(\d{2}/\d{2}/\d{4} @ \d{2}:\d{2}\)$`, expectedNick))
)

func Test_getTitleFromHTML(t *testing.T) {
	// --- --- --- --- --- --- Files with a title
	for _, fileName := range htmlsWithTitle {
		fileContent, err := ioutil.ReadFile(fileName)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		doc, err := html.Parse(strings.NewReader(string(fileContent)))
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		if title, found := getTitleFromHTML(doc); !found {
			t.Errorf("No title found in the document with title %q", expectedTitle)
		} else if title != expectedTitle {
			t.Errorf("Wrong title found: %q instead of %q", title, expectedTitle)
		}
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- File without a title
	fileContent, err := ioutil.ReadFile(htmlWithoutTitle)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	doc, err := html.Parse(strings.NewReader(string(fileContent)))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if title, found := getTitleFromHTML(doc); found {
		t.Errorf("Title found in the document without title: %q", title)
	}
	// --- --- --- --- --- ---
}

func Test_findURLs(t *testing.T) {
	// --- --- --- --- --- --- Messages with URLs
	for i, message := range messagesWithUrls {
		results := findURLs(message)
		if len(results) != len(expectedUrls[i]) {
			t.Errorf("Number of results different than what was expected (results: %d, expected: %d)", len(results), len(expectedUrls[i]))
		}
		for j, result := range results {
			if result.String() != expectedUrls[i][j] {
				t.Errorf("Url not matching: should be %q, is %q", expectedUrls[i][j], result.String())
			}
		}
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- Message without any URL
	if len(findURLs(messageWithoutURL)) != 0 {
		t.Errorf("Found URL(s) in the message without any URL (Message %q)", messageWithoutURL)
	}
	// --- --- --- --- --- ---
}

func Test_HandleURLs(t *testing.T) {
	db := database.NewDatabase("./tests.sqlite", "../../database/migrations", true)
	defer db.Close()
	Init(db)

	// --- --- --- --- --- --- Valid Event
	var testReply core.ReplyCallbackData
	HandleURLs(&validEvent, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- Invalid Event
	HandleURLs(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain any URL (Message: %q)\n\n", messageWithoutURL)
	})
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- Valid Event => Trigger the "link already posted" function
	var testReplies []core.ReplyCallbackData
	HandleURLs(&validEvent, func(data *core.ReplyCallbackData) {
		testReplies = append(testReplies, *data)
	})
	if len(testReplies) != 2 {
		t.Errorf("The test should trigger 2 callbacks, instead it triggered %d", len(testReplies))
	}

	// First reply is the "Link already posted" message
	if !re.MatchString(testReplies[0].Message) {
		t.Errorf("Regexp %q not matching %q", re.String(), testReplies[0].Message)
	}

	// Second reply is the page title
	if testReplies[1] != validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReplies[1], validReply)
	}
	// --- --- --- --- --- ---
}
