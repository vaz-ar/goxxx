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
	MsgHandlers       []func(*irc.Event, func(*ReplyCallbackData))
	CmdHandlers       []func(*irc.Event, func(*ReplyCallbackData)) bool
	MsgReplyCallbacks []func(*ReplyCallbackData)
	CmdReplyCallbacks []func(*ReplyCallbackData)
}

type ReplyCallbackData struct {
	Message string
	Nick    string
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
func (bot *Bot) AddMsgHandler(msgProcessCallback func(*irc.Event, func(*ReplyCallbackData)), replyCallback func(*ReplyCallbackData)) {
	if msgProcessCallback != nil && replyCallback != nil {
		bot.MsgHandlers = append(bot.MsgHandlers, msgProcessCallback)
		bot.MsgReplyCallbacks = append(bot.MsgReplyCallbacks, replyCallback)
	}
}

// cmdProcessCallback will be called on every user message the bot reads (if a command was not found previously in the message).
// replyCallback is to be called by cmdProcessCallback (or not) to yield and process its result as a string message.
// cmdProcessCallback must check if their replyCallback is nil before using it
// Command handlers must return true if they found a command to process, false otherwise
func (bot *Bot) AddCmdHandler(cmdProcessCallback func(*irc.Event, func(*ReplyCallbackData)) bool, replyCallback func(*ReplyCallbackData)) {
	if cmdProcessCallback != nil {
		bot.CmdHandlers = append(bot.CmdHandlers, cmdProcessCallback)
		bot.CmdReplyCallbacks = append(bot.CmdReplyCallbacks, replyCallback)
	}
}

func (bot *Bot) Run() {
	bot.ircConn.Loop()
}

func (bot *Bot) ReplyToAll(data *ReplyCallbackData) {
	bot.ircConn.Privmsg(bot.Channel, data.Message)
}

func (bot *Bot) ReplyToNick(data *ReplyCallbackData) {
	if data.Nick != "" {
		bot.ircConn.Privmsg(data.Nick, data.Message)
	}
}

func (bot *Bot) Reply(data *ReplyCallbackData) {
	if data.Nick == "" {
		bot.ircConn.Privmsg(bot.Channel, data.Message)
	} else {
		bot.ircConn.Privmsg(data.Nick, data.Message)
	}
}

func (bot *Bot) mainHandler(event *irc.Event) {
	for i, handler := range bot.CmdHandlers {
		if handler(event, bot.CmdReplyCallbacks[i]) {
			return
		}
	}
	for i, handler := range bot.MsgHandlers {
		handler(event, bot.MsgReplyCallbacks[i])
	}
}
