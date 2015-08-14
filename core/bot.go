// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.

// Contains the bot's core functionalities
package core

import (
	"github.com/thoj/go-ircevent"
)

// Bot structure that contains connection informations, IRC connection, command handlers and message handlers
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

// Structure used by the handlers to send data in a standardized format
type ReplyCallbackData struct {
	Message string
	Nick    string
}

// Bot constructor.
// Set the required parameters and open the connection to the server.
func NewBot(nick, server, channel, channelKey string) *Bot {
	bot := Bot{Nick: nick, Server: server, Channel: channel, ChannelKey: channelKey}
	bot.ircConn = irc.IRC(nick, nick)
	bot.ircConn.UseTLS = true
	bot.ircConn.Connect(server)
	bot.ircConn.Join(channel + " " + channelKey)
	bot.ircConn.AddCallback("PRIVMSG", bot.mainHandler)
	return &bot
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

// Start the event loop
func (bot *Bot) Run() {
	bot.ircConn.Loop()
}

// Send a message to the channel where the bot is connected
func (bot *Bot) ReplyToAll(data *ReplyCallbackData) {
	bot.ircConn.Privmsg(bot.Channel, data.Message)
}

// Send a private message to the user "data.Nick".
// If data.Nick is an empty string, do nothing
func (bot *Bot) ReplyToNick(data *ReplyCallbackData) {
	if data.Nick != "" {
		bot.ircConn.Privmsg(data.Nick, data.Message)
	}
}

// Send a private message to the user "data.Nick" if "data.Nick" isn't an empty string.
// If "data.Nick" is an empty string then send the message to the channel where the bot is connected.
func (bot *Bot) Reply(data *ReplyCallbackData) {
	if data.Nick == "" {
		bot.ircConn.Privmsg(bot.Channel, data.Message)
	} else {
		bot.ircConn.Privmsg(data.Nick, data.Message)
	}
}

// Handler called on every message posted in the channel where the bot is connected or directly sent to the bot
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
