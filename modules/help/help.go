// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Package help manages the help messages
package help

import (
	"fmt"
	"github.com/emirozer/go-helpers"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"strings"
)

const (
	defaultMessage = "You need to specify a module for which you want help. Currently loaded modules are %q."
)

var (
	helpMessages = map[string][]string{}
	modules      []string
)

// AddMessages stores messages to display them later via the help command
func AddMessages(cmd *core.Command) {
	if cmd.Module == "" || cmd.HelpMessage == "" {
		return
	}
	helpMessages[cmd.Module] = append(helpMessages[cmd.Module], cmd.HelpMessage)
	if !helpers.StringInSlice(cmd.Module, modules) {
		modules = append(modules, cmd.Module)
	}
}

// GetCommand returns a Command structure for the help command
func GetCommand() *core.Command {
	return &core.Command{
		Triggers: []string{"!h", "!help"},
		Handler:  handleHelpCmd}
}

// handleHelpCmd handles the !help command
func handleHelpCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1] => module
	if len(fields) != 2 {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf(defaultMessage, strings.Join(modules, ", ")), Target: event.Nick})
		return true
	}
	list, ok := helpMessages[fields[1]]
	if !ok {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf(defaultMessage, strings.Join(modules, ", ")), Target: event.Nick})
		return true
	}
	for _, message := range list {
		callback(&core.ReplyCallbackData{Message: message, Target: event.Nick})
	}
	return true
}
