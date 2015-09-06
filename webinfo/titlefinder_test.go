// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
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
	htmlWithTitle    string = "./tests_data/page_with_title.html"
	htmlWithoutTitle string = "./tests_data/page_without_title.html"
	expectedTitle    string = "Unicorn"
	expectedNick     string = "Sender"

	messagesWithUrls []string = []string{
		"Oh look a this link http://www.matmartinez.net/nsfw/ (It is not NSFW)",
		"https://golang.org/doc/effective_go.html", // -- 1
		"Check this one https://diasporafoundation.org/, and this other one cozy.io/en/, and this last one: http://framasoft.net/"}
	expectedUrls [][]string = [][]string{
		[]string{"http://www.matmartinez.net/nsfw/"},
		[]string{"https://golang.org/doc/effective_go.html"},
		[]string{"https://diasporafoundation.org/", "http://cozy.io/en/", "http://framasoft.net/"}}

	messageWithoutUrl string = "This is just.a.message without/any URL in.it"

	validEvent irc.Event = irc.Event{
		Nick:      expectedNick,
		Arguments: []string{"#test_channel", messagesWithUrls[1]}} // -- 1

	invalidEvent irc.Event = irc.Event{
		Nick:      expectedNick,
		Arguments: []string{"#test_channel", messageWithoutUrl}}

	validReply core.ReplyCallbackData = core.ReplyCallbackData{Nick: "", Message: "Effective Go - The Go Programming Language"} // -- 1

	re *regexp.Regexp = regexp.MustCompile(fmt.Sprintf(`^Link already posted by %s \(\d{2}/\d{2}/\d{4} @ \d{2}:\d{2}\)$`, expectedNick))
)

func Test_getTitleFromHTML(t *testing.T) {
	// --- --- --- --- --- --- File with a title
	fileContent, err := ioutil.ReadFile(htmlWithTitle)
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
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- File without a title
	fileContent, err = ioutil.ReadFile(htmlWithoutTitle)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	doc, err = html.Parse(strings.NewReader(string(fileContent)))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if title, found := getTitleFromHTML(doc); found {
		t.Errorf("Title found in the document without title: %q", title)
	}
	// --- --- --- --- --- ---
}

func Test_findUrls(t *testing.T) {
	// --- --- --- --- --- --- Messages with URLs
	for i, message := range messagesWithUrls {
		results := findUrls(message)
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
	if len(findUrls(messageWithoutUrl)) != 0 {
		t.Errorf("Found URL(s) in the message without any URL (Message %q)", messageWithoutUrl)
	}
	// --- --- --- --- --- ---
}

func Test_HandleUrls(t *testing.T) {
	db := database.NewDatabase("./tests.sqlite", true)
	defer db.Close()
	Init(db)

	// --- --- --- --- --- --- Valid Event
	var testReply core.ReplyCallbackData
	HandleUrls(&validEvent, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != validReply {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, validReply)
	}
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- Invalid Event
	HandleUrls(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain any URL (Message: %q)\n\n", messageWithoutUrl)
	})
	// --- --- --- --- --- ---

	// --- --- --- --- --- --- Valid Event => Trigger the "link already posted" function
	var testReplies []core.ReplyCallbackData
	HandleUrls(&validEvent, func(data *core.ReplyCallbackData) {
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
