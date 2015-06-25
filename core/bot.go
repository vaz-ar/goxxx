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
	Nick       string
	Server     string
	Channel    string
	ChannelKey string
	ircConn    *irc.Connection
}

func (bot *Bot) Init() {
	bot.ircConn = irc.IRC(bot.Nick, bot.Nick)
	bot.ircConn.UseTLS = true
	bot.ircConn.Connect(bot.Server)
	bot.ircConn.Join(bot.Channel + " " + bot.ChannelKey)
}

// msgProcessCallback will be called on every user message the bot reads.
// replyCallback is to be called by msgProcessCallback (or not) to yield
// and process its result as a string message.
func (bot *Bot) AddMsgHandler(msgProcessCallback func(string, func(string)), replyCallback func(string)) {
	bot.ircConn.AddCallback("PRIVMSG", func(event *irc.Event) {
		msgProcessCallback(event.Message(), replyCallback)
	})
}

func (bot *Bot) Run() {
	bot.ircConn.Loop()
}

func (bot *Bot) ReplyToAll(message string) {
	bot.ircConn.Privmsg(bot.Channel, message)
}
