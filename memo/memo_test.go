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

var (
	validMessage   string = "  \t  !memo   Receiver this is a memo      "
	invalidMessage string = "this is not a memo command"
	expectedNick   string = "Receiver"

	// For the Arguments field I checked how it worked from a real function call (Not documented)
	validEvent irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", validMessage}}

	invalidEvent irc.Event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", invalidMessage}}

	replyCallbackDataReference core.ReplyCallbackData = core.ReplyCallbackData{Nick: "Sender", Message: "Sender: memo for Receiver saved"}
)

func Test_HandleMemoCmd(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	db := database.InitDatabase("tests.sqlite", true)
	defer db.Close()
	Init(db)

	// --- --- --- Supposed to pass
	var replyCallbackDataTest core.ReplyCallbackData
	HandleMemoCmd(&validEvent, func(data *core.ReplyCallbackData) {
		replyCallbackDataTest = *data
	})
	if replyCallbackDataTest != replyCallbackDataReference {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", replyCallbackDataTest, replyCallbackDataReference)
	}
	// --- --- --- --- --- ---

	// --- --- --- Not supposed to pass
	HandleMemoCmd(&invalidEvent, func(data *core.ReplyCallbackData) {
		// There is no memo command in the message, the callback should not be called
		t.Errorf("Callback function not supposed to be called, the message does not contain the !memo command (Message: %q)\n\n", invalidMessage)
	})
	// --- --- --- --- --- ---
}

func Test_SendMemo(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	db := database.InitDatabase("tests.sqlite", true)
	defer db.Close()
	Init(db)

	// Create Memo
	HandleMemoCmd(&validEvent, nil)

	message := " this is a message to trigger the memo "
	event := irc.Event{Nick: expectedNick, Arguments: []string{"#test_channel", message}}
	re := regexp.MustCompile(fmt.Sprintf(`^%s: memo from Sender => "this is a memo" \(\d{2}/\d{2}/\d{4} @ \d{2}:\d{2}\)$`, expectedNick))

	var replyCallbackDataTest core.ReplyCallbackData
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
