// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Package help manages the help messages
package help

import (
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"strings"
)

var messageList []string

// AddMessages stores messages to display them later via the help command
func AddMessages(helpMessages ...string) {
	for _, message := range helpMessages {
		messageList = append(messageList, message)
	}
}

// HandleHelpCmd handles the !help command
func HandleHelpCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	if len(fields) == 0 || fields[0] != "!help" {
		return false
	}
	for _, message := range messageList {
		callback(&core.ReplyCallbackData{Message: message, Nick: event.Nick})
	}
	return true
}
