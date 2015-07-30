// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package memo

import (
	"fmt"
	// "github.com/fatih/color"
	"github.com/romainletendart/goxxx/core"
	"github.com/romainletendart/goxxx/database"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"regexp"
	"testing"
)

func TestHandleMemoCmd(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	db := database.InitDatabase("tests.sqlite", true)
	defer db.Close()
	Init(db)

	// --- --- --- Supposed to pass
	var (
		message string = "  \t  !memo Receiver this is a memo      "

		// For the Arguments field I checked how it worked from a real function call (Not documented)
		event irc.Event = irc.Event{
			Nick:      "Sender",
			Arguments: []string{"#test_channel", message}}

		replyCallbackDataTest      core.ReplyCallbackData
		replyCallbackDataReference core.ReplyCallbackData = core.ReplyCallbackData{Nick: "Sender", Message: "Sender: memo for Receiver saved"}
	)

	HandleMemoCmd(&event, func(data *core.ReplyCallbackData) {
		replyCallbackDataTest = *data
	})

	if replyCallbackDataTest != replyCallbackDataReference {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", replyCallbackDataTest, replyCallbackDataReference)
	}
	// --- --- --- --- --- ---

	// --- --- --- Not supposed to pass
	message = " this is not a command "
	event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", message}}

	// There is no memo command in the message, the callback should not be called
	HandleMemoCmd(&event, func(data *core.ReplyCallbackData) {
		t.Errorf("Callback function not supposed to be called, the message does not contain the !memo command (Message: %q)\n\n", message)
	})
	// --- --- --- --- --- ---
}

func TestSendMemo(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	db := database.InitDatabase("tests.sqlite", true)
	defer db.Close()
	Init(db)

	var (
		message               string    = "!memo Receiver this is a memo"
		expectedNick          string    = "Receiver"
		event                 irc.Event = irc.Event{Nick: "Sender", Arguments: []string{"#test_channel", message}}
		replyCallbackDataTest core.ReplyCallbackData
	)

	// Create Memo
	HandleMemoCmd(&event, nil)

	message = " this is a message to trigger the memo "
	event = irc.Event{Nick: expectedNick, Arguments: []string{"#test_channel", message}}
	re := regexp.MustCompile(fmt.Sprintf(`^%s: memo from Sender => "this is a memo" \(\d{2}/\d{2}/\d{4} @ \d{2}:\d{2}\)$`, expectedNick))

	SendMemo(&event, func(data *core.ReplyCallbackData) {
		replyCallbackDataTest = *data
	})

	if !re.MatchString(replyCallbackDataTest.Message) {
		t.Errorf("Regexp %q not matching %q", re.String(), replyCallbackDataTest.Message)
	}
	if replyCallbackDataTest.Nick != expectedNick {
		t.Errorf("Incorrect Nick: should be %q, is %q", expectedNick, replyCallbackDataTest.Nick)
	}
}
