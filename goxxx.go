// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.

// Main package for the goxxx project
//
// For the details see the file goxxx.go, as godoc won't show the documentation for the main package
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

// Process the command line arguments
func getOptions() (nick, server, channel, channelKey string, debug bool, returnCode int) {
	flag.StringVar(&channel, "channel", "", "IRC channel name")
	flag.StringVar(&channelKey, "key", "", "IRC channel key (optional)")
	flag.StringVar(&nick, "nick", "goxxx", "the bot's nickname (optional)")
	flag.StringVar(&server, "server", "chat.freenode.net:6697", "IRC_SERVER[:PORT] (optional)")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
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

	nick, server, channel, channelKey, debug, returnCode := getOptions()
	if returnCode == FLAGS_EXIT {
		return
	} else if returnCode == FLAGS_FAILURE {
		log.Fatal("Initialisation failed (getOptions())")
	}
	if debug {
		// In debug mode we show the file name and the line from where the log come from
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create the database
	db := database.NewDatabase("", false)
	defer db.Close()

	// Create the bot
	bot := core.NewBot(nick, server, channel, channelKey)

	// Initialise packages
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

	// Message Handlers
	bot.AddMsgHandler(webinfo.HandleUrls, bot.ReplyToAll)
	bot.AddMsgHandler(memo.SendMemo, bot.ReplyToNick)

	// Command Handlers
	bot.AddCmdHandler(memo.HandleMemoCmd, bot.ReplyToAll)
	bot.AddCmdHandler(memo.HandleMemoStatusCmd, bot.ReplyToNick)
	bot.AddCmdHandler(search.HandleSearchCmd, bot.Reply)
	bot.AddCmdHandler(help.HandleHelpCmd, bot.ReplyToNick)
	bot.AddCmdHandler(xkcd.HandleXKCDCmd, bot.ReplyToAll)

	// Start the bot
	bot.Run()
}
