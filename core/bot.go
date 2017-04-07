// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.

// Package core contains the bot's core functionalities
package core

import (
	"github.com/emirozer/go-helpers"
	"github.com/thoj/go-ircevent"
	// "log"
	"strings"
	"time"
)

var (
	updateUserListDone = make(chan bool, 1)
)

// Bot structure that contains connection informations, IRC connection, command handlers and message handlers
type Bot struct {
	nick              string
	server            string
	channel           string
	channelKey        string
	Admins            *[]string
	users             []string
	ircConn           *irc.Connection
	msgHandlers       []func(*irc.Event, func(*ReplyCallbackData))
	msgReplyCallbacks []func(*ReplyCallbackData)
	cmdHandlers       map[string]func(*irc.Event, func(*ReplyCallbackData)) bool
	cmdReplyCallbacks map[string]func(*ReplyCallbackData)
	lastReplyTime     time.Time
}

// ReplyCallbackData Structure used by the handlers to send data in a standardized format
type ReplyCallbackData struct {
	Message string // Message to send
	Target  string // Destination target of the message (Channel or Nick)
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
	bot := Bot{nick: nick, server: server, channel: channel, channelKey: channelKey, Admins: new([]string)}

	bot.ircConn = irc.IRC(nick, nick)
	bot.ircConn.UseTLS = true
	bot.ircConn.Connect(server)

	//PRIVMSG
	bot.ircConn.AddCallback("PRIVMSG", bot.mainHandler)

	// RPL_WELCOME
	bot.ircConn.AddCallback("001", func(event *irc.Event) {
		go func(event *irc.Event) {
			bot.ircConn.Join(channel + " " + channelKey)
			// Necessary because the callback for RPL_NAMREPLY is called after joining the channel (NAMES command)
			// If not called here updateUserListDone will always contains a value before being read by UpdateUserList()
			<-updateUserListDone
		}(event)
	})

	// RPL_NAMREPLY
	bot.ircConn.AddCallback("353", func(event *irc.Event) {
		var (
			currentAdmins []string
			currentUsers  []string
		)
		for _, user := range strings.Split(event.Message(), " ") {
			if strings.HasPrefix(user, "@") {
				currentAdmins = append(currentAdmins, strings.TrimPrefix(user, "@"))
				currentUsers = append(currentUsers, strings.TrimPrefix(user, "@"))
			} else {
				currentUsers = append(currentUsers, user)
			}
		}
		*bot.Admins = currentAdmins
		bot.users = currentUsers
		// log.Printf("Current admnistrators: %s", strings.Join(*bot.Admins, ", "))
		// log.Printf("Current users: %s", strings.Join(bot.users, ", "))
		updateUserListDone <- true
	})

	bot.cmdHandlers = make(map[string]func(*irc.Event, func(*ReplyCallbackData)) bool)
	bot.cmdReplyCallbacks = make(map[string]func(*ReplyCallbackData))
	bot.lastReplyTime = time.Now()

	return &bot
}

// AddMsgHandler adds a message handler to bot.
// msgProcessCallback will be called on every user message the bot reads (if a command was not found previously in the message).
// replyCallback is to be called by msgProcessCallback (or not) to yield and process its result as a string message.
func (bot *Bot) AddMsgHandler(msgProcessCallback func(*irc.Event, func(*ReplyCallbackData)), replyCallback func(*ReplyCallbackData)) {
	if msgProcessCallback != nil {
		bot.msgHandlers = append(bot.msgHandlers, msgProcessCallback)
		bot.msgReplyCallbacks = append(bot.msgReplyCallbacks, replyCallback)
	}
}

// AddCmdHandler adds a command handler to bot.
// cmdStruct is a pointer to a Command structure.
// replyCallback is to be called by cmdProcessCallback (or not) to yield and process its result as a string message.
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
	bot.reply(bot.channel, data.Message)
}

// Reply sends a message to the user or channel specifed by "data.Target".
func (bot *Bot) Reply(data *ReplyCallbackData) {
	if data.Target != "" {
		bot.reply(data.Target, data.Message)
	}
}

// reply sends a message and introduces necessary pauses between consecutive messages to deal with flood control
func (bot *Bot) reply(target string, message string) {
	elapsedTime := time.Since(bot.lastReplyTime)
	if elapsedTime < (2 * time.Second) {
		time.Sleep((2 * time.Second) - elapsedTime)
	}
	bot.ircConn.Privmsg(target, message)
	bot.lastReplyTime = time.Now()
}

// mainHandler is called on every message posted in the channel where the bot is connected or directly sent to the bot.
func (bot *Bot) mainHandler(event *irc.Event) {

	if strings.TrimSpace(event.Message()) == "" {
		return
	}

	cmd := strings.Fields(event.Message())[0]
	cmdHandler, present := bot.cmdHandlers[cmd]
	if present {
		go func() {
			if !helpers.StringInSlice(event.Nick, bot.users) {
				UpdateUserList(event)
				if !helpers.StringInSlice(event.Nick, bot.users) {
					return
				}
			}
			go cmdHandler(event, bot.cmdReplyCallbacks[cmd])
		}()
	}

	for i, handler := range bot.msgHandlers {
		go handler(event, bot.msgReplyCallbacks[i])
	}
}

// UpdateUserList Update the user list used for access control
func UpdateUserList(event *irc.Event) {
	event.Connection.SendRawf("NAMES %s", event.Arguments[0])
	<-updateUserListDone
}

// GetTargetFromEvent If the message originated from a channel then return it, else return the nick that sent the message
func GetTargetFromEvent(event *irc.Event) string {
	source := strings.TrimSpace(event.Arguments[0])
	if strings.HasPrefix(source, "#") {
		return source
	}
	return event.Nick
}
