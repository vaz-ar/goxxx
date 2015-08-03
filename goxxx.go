// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.
package main

import (
	"flag"
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/romainletendart/goxxx/database"
	"github.com/romainletendart/goxxx/memo"
	"github.com/romainletendart/goxxx/search"
	"github.com/romainletendart/goxxx/webinfo"
	"log"
	"os"
)

const (
	// Application version
	global_version string = "1.0.0"

	// Equivalent to enums (cf. https://golang.org/ref/spec#Iota)
	flags_exit    = iota //  == 0
	flags_success        //  == 1
	flags_failure        //  == 2
)

func getOptions() (nick, server, channel, channelKey string, returnCode int) {
	flag.StringVar(&channel, "channel", "", "IRC channel name")
	flag.StringVar(&channelKey, "key", "", "IRC channel key (optional)")
	flag.StringVar(&nick, "nick", "goxxx", "the bot's nickname (optional)")
	flag.StringVar(&server, "server", "chat.freenode.net:6697", "IRC_SERVER[:PORT] (optional)")
	version := flag.Bool("version", false, "Display goxxx version")

	flag.Usage = func() {
		fmt.Println("Usage:", os.Args[0], "-channel CHANNEL [ARGUMENTS]")
		fmt.Println()
		fmt.Println("Arguments description:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *version {
		fmt.Printf("\nGoxxx version: %s\n\n", global_version)
		returnCode = flags_exit
		return
	}

	if channel == "" {
		flag.Usage()
		returnCode = flags_failure
	} else {
		returnCode = flags_success
	}

	return
}

func main() {

	// Set log output to a file
	logFile, err := os.OpenFile("./logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	nick, server, channel, channelKey, returnCode := getOptions()
	if returnCode == flags_exit {
		return
	} else if returnCode == flags_failure {
		log.Fatal("Initialisation failed (getOptions())")
	}

	db := database.InitDatabase("", false)
	defer db.Close()

	bot := core.Bot{
		Nick:       nick,
		Server:     server,
		Channel:    channel,
		ChannelKey: channelKey,
	}
	bot.Init()
	memo.Init(db)
	webinfo.Init(db)
	search.Init()

	bot.AddMsgHandler(webinfo.HandleUrls, bot.ReplyToAll)
	bot.AddMsgHandler(memo.SendMemo, bot.ReplyToNick)

	bot.AddCmdHandler(memo.HandleMemoCmd, bot.ReplyToAll)
	bot.AddCmdHandler(memo.HandleMemoStatusCmd, bot.ReplyToNick)
	bot.AddCmdHandler(search.HandleSearchCmd, bot.ReplyToAll)

	bot.Run()
}
