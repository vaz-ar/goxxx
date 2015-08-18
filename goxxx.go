// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.

// Main package for the goxxx project
//
// For the details see the file goxxx.go, as godoc won't show the documentation for the main package.
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
	"github.com/vharitonsky/iniflags"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Hybrid config: use flags and INI file
	// Command line flags take precedence on INI values
	// INI file path can be passed via the -command flag or via a function (commented for now, exit application if the file does not exist ...)
	//     iniflags.SetConfigFile("./config.ini")
	iniflags.Parse()

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

	// Go signal notification works by sending os.Signal values on a channel.
	// We'll create a channel to receive these notifications
	// (we'll also make one to notify us when the program can exit).
	interruptSignals := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// signal.Notify registers the given channel to receive notifications of the specified signals.
	signal.Notify(interruptSignals, syscall.SIGINT, syscall.SIGTERM)

	// This goroutine executes a blocking receive for signals.
	// When it gets one it'll print it out and then notify the program that it can finish.
	go func() {
		sig := <-interruptSignals
		log.Printf("\nSystem signal received: %s", sig)
		done <- true
	}()

	// Start the bot
	go bot.Run()

	// The current routine will be blocked here until done is true
	<-done

	// Close the bot connection and the database
	bot.Stop()
	db.Close()

	log.Println("Goxxx exiting")
}
