// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.
package core

import (
	"github.com/thoj/go-ircevent"
)

type Bot struct {
	Nick              string
	Server            string
	Channel           string
	ChannelKey        string
	ircConn           *irc.Connection
	msgHandlers       []func(*irc.Event, func(string))
	cmdHandlers       []func(*irc.Event, func(string)) bool
	msgReplyCallbacks []func(string)
	cmdReplyCallbacks []func(string)
}

func (bot *Bot) Init() {
	bot.ircConn = irc.IRC(bot.Nick, bot.Nick)
	bot.ircConn.UseTLS = true
	bot.ircConn.Connect(bot.Server)
	bot.ircConn.Join(bot.Channel + " " + bot.ChannelKey)
	bot.ircConn.AddCallback("PRIVMSG", bot.mainHandler)
}

// msgProcessCallback will be called on every user message the bot reads (if a command was not found previously in the message).
// replyCallback is to be called by msgProcessCallback (or not) to yield and process its result as a string message.
func (bot *Bot) AddMsgHandler(msgProcessCallback func(*irc.Event, func(string)), replyCallback func(string)) {
	if msgProcessCallback != nil && replyCallback != nil {
		bot.msgHandlers = append(bot.msgHandlers, msgProcessCallback)
		bot.msgReplyCallbacks = append(bot.msgReplyCallbacks, replyCallback)
	}
}

// cmdProcessCallback will be called on every user message the bot reads (if a command was not found previously in the message).
// replyCallback is to be called by cmdProcessCallback (or not) to yield and process its result as a string message.
// cmdProcessCallback must check if their replyCallback is nil before using it
// Command handlers must return true if they found a command to process, false otherwise
func (bot *Bot) AddCmdHandler(cmdProcessCallback func(*irc.Event, func(string)) bool, replyCallback func(string)) {
	if cmdProcessCallback != nil {
		bot.cmdHandlers = append(bot.cmdHandlers, cmdProcessCallback)
		bot.cmdReplyCallbacks = append(bot.cmdReplyCallbacks, replyCallback)
	}
}

func (bot *Bot) Run() {
	bot.ircConn.Loop()
}

func (bot *Bot) ReplyToAll(message string) {
	bot.ircConn.Privmsg(bot.Channel, message)
}

func (bot *Bot) mainHandler(event *irc.Event) {
	for i, handler := range bot.cmdHandlers {
		if handler(event, bot.cmdReplyCallbacks[i]) {
			return
		}
	}
	for i, handler := range bot.msgHandlers {
		handler(event, bot.msgReplyCallbacks[i])
	}
}
