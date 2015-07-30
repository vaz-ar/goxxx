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

func getOptions() (nick, server, channel, channelKey string, success bool) {
	flag.StringVar(&channel, "channel", "", "IRC channel name")
	flag.StringVar(&channelKey, "key", "", "IRC channel key (optional)")
	flag.StringVar(&nick, "nick", "goxxx", "the bot's nickname (optional)")
	flag.StringVar(&server, "server", "chat.freenode.net:6697", "IRC_SERVER[:PORT] (optional)")
	flag.Usage = func() {
		fmt.Println("Usage:", os.Args[0], "-channel CHANNEL [ARGUMENTS]")
		fmt.Println()
		fmt.Println("Arguments description:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if channel == "" {
		flag.Usage()
		success = false
	} else {
		success = true
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

	nick, server, channel, channelKey, success := getOptions()
	if !success {
		log.Fatal("Initialisation failed (getOptions())")
		return
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
	bot.AddMsgHandler(memo.SendMemo, bot.ReplyToAll)

	bot.AddCmdHandler(memo.HandleMemoCmd, bot.ReplyToAll)
	bot.AddCmdHandler(memo.HandleMemoStatusCmd, bot.ReplyToNick)
	bot.AddCmdHandler(search.HandleSearchCmd, bot.ReplyToAll)

	bot.Run()
}
