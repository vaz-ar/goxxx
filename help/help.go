// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Package help manages the help messages
package help

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"strings"
)

const (
	defaultMessage = "You need to specify a module for which you want help. Currently loaded modules are %q."
)

var (
	helpMessages = map[string][]string{}
	modules      string
)

// AddMessages stores messages to display them later via the help command
func AddMessages(keyword string, messages ...string) {
	if keyword == "" || len(messages) == 0 {
		return
	}
	helpMessages[keyword] = messages
	if modules == "" {
		modules = keyword
	} else {
		modules += ", " + keyword
	}
}

// HandleHelpCmd handles the !help command
func HandleHelpCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1] => module
	if len(fields) < 1 || fields[0] != "!help" {
		return false
	} else if len(fields) != 2 {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf(defaultMessage, modules), Nick: event.Nick})
		return true
	}
	list, ok := helpMessages[fields[1]]
	if !ok {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf(defaultMessage, modules), Nick: event.Nick})
		return true
	}
	for _, message := range list {
		callback(&core.ReplyCallbackData{Message: message, Nick: event.Nick})
	}
	return true
}
