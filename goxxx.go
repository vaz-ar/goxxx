// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.
package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/thoj/go-ircevent"
)

func getOptions() (nick, server, channel, channelKey string, success bool) {
	flag.StringVar(&channel, "channel", "", "IRC channel name")
	flag.StringVar(&channelKey, "key", "", "IRC channel key (optional)")
	flag.StringVar(&nick, "nick", "goxxx", "the bot's nickname (optional)")
	flag.StringVar(&server, "server", "chat.freenode.net:6697", "IRC_SERVER[:PORT] (optional)")
	flag.Parse()

	if channel == "" {
		fmt.Println("Usage:", os.Args[0], "-channel CHANNEL [ARGUMENTS]")
		fmt.Println()
		fmt.Println("Arguments description:")
		flag.PrintDefaults()
		success = false
	} else {
		success = true
	}

	return
}

func main() {
	nick, server, channel, channelKey, success := getOptions()

	if !success {
		return
	}

	ircConn := irc.IRC(nick, nick)
	ircConn.UseTLS = true
	ircConn.Connect(server)
	ircConn.Join(channel + " " +  channelKey)
	ircConn.Privmsg(channel, "Hello world")

	ircConn.Loop()
}
