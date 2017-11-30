// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
//
// See LICENSE file.

/*
Package quote contains quote commands
*/
package core

import (
	"github.com/thoj/go-ircevent"
	"testing"
)

func Test_GetTargetFromEvent(t *testing.T) {

	// Channel
	event := irc.Event{
		Nick:      "Sender",
		Arguments: []string{"#test_channel", "test message"}}
	expectedResult := "#test_channel"

	result := GetTargetFromEvent(&event)

	if result != expectedResult {
		t.Errorf("Result not matching expected result (%q != %q)", result, expectedResult)
	}

	// Nick
	event = irc.Event{
		Nick:      "Sender",
		Arguments: []string{"Receiver", "test message"}}
	expectedResult = "Sender"

	result = GetTargetFromEvent(&event)

	if result != expectedResult {
		t.Errorf("Result not matching expected result (%q != %q)", result, expectedResult)
	}
}
