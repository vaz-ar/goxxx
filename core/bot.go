// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.

// Package core contains the bot's core functionalities
package core

import (
	"github.com/thoj/go-ircevent"
	"strings"
)

// Bot structure that contains connection informations, IRC connection, command handlers and message handlers
type Bot struct {
	nick              string
	server            string
	channel           string
	channelKey        string
	ircConn           *irc.Connection
	msgHandlers       []func(*irc.Event, func(*ReplyCallbackData))
	msgReplyCallbacks []func(*ReplyCallbackData)
	cmdHandlers       map[string]func(*irc.Event, func(*ReplyCallbackData)) bool
	cmdReplyCallbacks map[string]func(*ReplyCallbackData)
}

// ReplyCallbackData Structure used by the handlers to send data in a standardized format
type ReplyCallbackData struct {
	Message string
	Nick    string
}

// Command structure
type Command struct {
	Module      string
	HelpMessage string
	Triggers    []string
	Handler     func(event *irc.Event, callback func(*ReplyCallbackData)) bool
}

// NewBot creates a new Bot, sets the required parameters and open the connection to the server.
func NewBot(nick, server, channel, channelKey string) *Bot {
	bot := Bot{nick: nick, server: server, channel: channel, channelKey: channelKey}
	bot.ircConn = irc.IRC(nick, nick)
	bot.ircConn.UseTLS = true
	bot.ircConn.Connect(server)
	bot.ircConn.Join(channel + " " + channelKey)
	bot.ircConn.AddCallback("PRIVMSG", bot.mainHandler)
	bot.cmdHandlers = make(map[string]func(*irc.Event, func(*ReplyCallbackData)) bool)
	bot.cmdReplyCallbacks = make(map[string]func(*ReplyCallbackData))
	return &bot
}

// AddMsgHandler adds a message handler to bot.
// msgProcessCallback will be called on every user message the bot reads (if a command was not found previously in the message).
// replyCallback is to be called by msgProcessCallback (or not) to yield and process its result as a string message.
func (bot *Bot) AddMsgHandler(msgProcessCallback func(*irc.Event, func(*ReplyCallbackData)), replyCallback func(*ReplyCallbackData)) {
	if msgProcessCallback != nil && replyCallback != nil {
		bot.msgHandlers = append(bot.msgHandlers, msgProcessCallback)
		bot.msgReplyCallbacks = append(bot.msgReplyCallbacks, replyCallback)
	}
}

// AddCmdHandler adds a command handler to bot.
// cmdProcessCallback will be called on every user message the bot reads (if a command was not found previously in the message).
// replyCallback is to be called by cmdProcessCallback (or not) to yield and process its result as a string message.
// cmdProcessCallback must check if their replyCallback is nil before using it
// Command handlers must return true if they found a command to process, false otherwise
func (bot *Bot) AddCmdHandler(cmdStruct *Command, replyCallback func(*ReplyCallbackData)) {
	if cmdStruct.Handler == nil {
		return
	}
	for _, command := range cmdStruct.Triggers {
		bot.cmdHandlers[command] = cmdStruct.Handler
		bot.cmdReplyCallbacks[command] = replyCallback
	}
}

// Run starts the event loop
func (bot *Bot) Run() {
	bot.ircConn.Loop()
}

// Stop exits the event loop
func (bot *Bot) Stop() {
	// Quit the current connection and disconnect from the server (details: https://tools.ietf.org/html/rfc1459#section-4.1.6)
	bot.ircConn.Quit()
}

// ReplyToAll sends a message to the channel where the bot is connected
func (bot *Bot) ReplyToAll(data *ReplyCallbackData) {
	bot.ircConn.Privmsg(bot.channel, data.Message)
}

// ReplyToNick sends a private message to the user "data.Nick".
// If data.Nick is an empty string, do nothing
func (bot *Bot) ReplyToNick(data *ReplyCallbackData) {
	if data.Nick != "" {
		bot.ircConn.Privmsg(data.Nick, data.Message)
	}
}

// Reply sends a private message to the user "data.Nick" if "data.Nick" isn't an empty string.
// If "data.Nick" is an empty string then send the message to the channel where the bot is connected.
func (bot *Bot) Reply(data *ReplyCallbackData) {
	if data.Nick == "" {
		bot.ircConn.Privmsg(bot.channel, data.Message)
	} else {
		bot.ircConn.Privmsg(data.Nick, data.Message)
	}
}

// mainHandler is called on every message posted in the channel where the bot is connected or directly sent to the bot.
func (bot *Bot) mainHandler(event *irc.Event) {

	cmd := strings.Fields(event.Message())[0]
	cmdHandler, present := bot.cmdHandlers[cmd]
	if present && cmdHandler(event, bot.cmdReplyCallbacks[cmd]) {
		return
	}

	for i, handler := range bot.msgHandlers {
		handler(event, bot.msgReplyCallbacks[i])
	}
}
