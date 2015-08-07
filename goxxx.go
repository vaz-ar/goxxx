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
	"github.com/romainletendart/goxxx/help"
	"github.com/romainletendart/goxxx/memo"
	"github.com/romainletendart/goxxx/search"
	"github.com/romainletendart/goxxx/webinfo"
	"github.com/romainletendart/goxxx/xkcd"
	"log"
	"os"
)

const (
	// Application version
	GLOBAL_VERSION string = "0.0.1"

	// Equivalent to enums (cf. https://golang.org/ref/spec#Iota)
	FLAGS_EXIT    = iota //  == 0
	FLAGS_SUCCESS        //  == 1
	FLAGS_FAILURE        //  == 2
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
		fmt.Printf("\nGoxxx version: %s\n\n", GLOBAL_VERSION)
		returnCode = FLAGS_EXIT
		return
	}

	if channel == "" {
		flag.Usage()
		returnCode = FLAGS_FAILURE
	} else {
		returnCode = FLAGS_SUCCESS
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
	if returnCode == FLAGS_EXIT {
		return
	} else if returnCode == FLAGS_FAILURE {
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

	help.Init(
		search.HELP_DUCKDUCKGO,
		search.HELP_WIKIPEDIA,
		search.HELP_WIKIPEDIA_FR,
		search.HELP_URBANDICTIONNARY,
		memo.HELP_MEMO,
		memo.HELP_MEMOSTAT,
		xkcd.HELP_XKCD,
		xkcd.HELP_XKCD_NUM)

	bot.AddMsgHandler(webinfo.HandleUrls, bot.ReplyToAll)
	bot.AddMsgHandler(memo.SendMemo, bot.ReplyToNick)

	bot.AddCmdHandler(memo.HandleMemoCmd, bot.ReplyToAll)
	bot.AddCmdHandler(memo.HandleMemoStatusCmd, bot.ReplyToNick)
	bot.AddCmdHandler(search.HandleSearchCmd, bot.Reply)
	bot.AddCmdHandler(help.HandleHelpCmd, bot.ReplyToNick)
	bot.AddCmdHandler(xkcd.HandleXKCDCmd, bot.ReplyToAll)

	bot.Run()
}
