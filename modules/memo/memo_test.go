// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
//
// See LICENSE file.
package memo

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"github.com/vaz-ar/goxxx/database"
	"regexp"
	"testing"
)

var (
	validMessage   = "  \t  !memo   Receiver this is a memo      "
	invalidMessage = "this is not a memo command"
	expectedNick   = "Receiver"

	// For the Arguments field I checked how it worked from a real function call (Not documented)
	validEvent = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", validMessage}}

	replyCallbackDataReference = core.ReplyCallbackData{Target: "Sender", Message: "Sender: memo for Receiver saved"}
)

func Test_handleMemoCmd(t *testing.T) {

	db := database.NewDatabase("./tests.sqlite", "../../database/migrations", true)
	defer db.Close()
	Init(db)

	// --- --- --- --- --- --- Valid Event
	var testReply core.ReplyCallbackData
	handleMemoCmd(&validEvent, func(data *core.ReplyCallbackData) {
		testReply = *data
	})
	if testReply != replyCallbackDataReference {
		t.Errorf("Test data differ from reference data:\nTest data:\t%#v\nReference data: %#v\n\n", testReply, replyCallbackDataReference)
	}
	// --- --- --- --- --- ---
}

func Test_SendMemo(t *testing.T) {

	db := database.NewDatabase("./tests.sqlite", "../../database/migrations", true)
	defer db.Close()
	Init(db)

	// Create Memo
	handleMemoCmd(&validEvent, nil)

	message := " this is a message to trigger the memo "
	event := irc.Event{Nick: expectedNick, Arguments: []string{"#test_channel", message}}
	re := regexp.MustCompile(fmt.Sprintf(`^%s: memo from Sender => "this is a memo" \(\d{2}/\d{2}/\d{4} @ \d{2}:\d{2}\)$`, expectedNick))

	var testReply core.ReplyCallbackData
	SendMemo(&event, func(data *core.ReplyCallbackData) {
		testReply = *data
	})

	if !re.MatchString(testReply.Message) {
		t.Errorf("Regexp %q not matching %q", re.String(), testReply.Message)
	}
	if testReply.Target != expectedNick {
		t.Errorf("Incorrect Nick: should be %q, is %q", expectedNick, testReply.Target)
	}
}
