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
	"github.com/vaz-ar/cfg_flags"
	"github.com/vaz-ar/goxxx/core"
	"github.com/vaz-ar/goxxx/database"
	"github.com/vaz-ar/goxxx/help"
	"github.com/vaz-ar/goxxx/invoke"
	"github.com/vaz-ar/goxxx/memo"
	"github.com/vaz-ar/goxxx/search"
	"github.com/vaz-ar/goxxx/webinfo"
	"github.com/vaz-ar/goxxx/xkcd"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	// Application version
	GLOBAL_VERSION string = "0.0.1"

	// Equivalent to enums (cf. https://golang.org/ref/spec#Iota)
	FLAGS_EXIT     = iota //  == 0
	FLAGS_SUCCESS         //  == 1
	FLAGS_FAILURE         //  == 2
	FLAGS_ADD_USER        //  == 3
)

// Config struct
type configData struct {
	channel       string
	channelKey    string
	nick          string
	server        string
	modules       []string
	debug         bool
	emailServer   string
	emailPort     int
	emailSender   string
	emailAccount  string
	emailPassword string
}

// Process the command line arguments
func getOptions() (config configData, returnCode int) {
	// IRC
	flag.StringVar(&config.channel, "channel", "", "IRC channel name")
	flag.StringVar(&config.channelKey, "key", "", "IRC channel key (optional)")
	flag.StringVar(&config.nick, "nick", "goxxx", "the bot's nickname (optional)")
	flag.StringVar(&config.server, "server", "chat.freenode.net:6697", "IRC_SERVER[:PORT] (optional)")
	modules := flag.String("modules", "memo,webinfo,invoke,search,xkcd", "Modules to enable (separated by commas)")
	// Email
	flag.StringVar(&config.emailServer, "email_server", "", "SMTP server address")
	flag.IntVar(&config.emailPort, "email_port", 0, "SMTP server port")
	flag.StringVar(&config.emailSender, "email_sender", "", "Email address to use in the \"From\" part of the header")
	flag.StringVar(&config.emailAccount, "email_account", "", "Email address from which to send emails")
	flag.StringVar(&config.emailPassword, "email_pwd", "", "password for the SMTP server")
	// Application
	flag.BoolVar(&config.debug, "debug", false, "Debug mode")
	version := flag.Bool("version", false, "Display goxxx version")

	flag.Usage = func() {
		fmt.Println("Usage:", os.Args[0], "-channel CHANNEL [ARGUMENTS]")
		fmt.Println()
		fmt.Println("Arguments description:")
		flag.PrintDefaults()
		fmt.Println("\nCommands description:")
		fmt.Println("  add_user <nick> <email>: Add an user to the database\n")
	}

	// Hybrid config: use flags and INI file
	// Command line flags take precedence on INI values
	if err := cfg_flags.Parse("goxxx.ini"); err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	config.modules = strings.Split(*modules, ",")

	if *version {
		fmt.Printf("\nGoxxx version: %s\n\n", GLOBAL_VERSION)
		returnCode = FLAGS_EXIT
		return
	}

	lenArgs := len(flag.Args())
	// add_user command
	if lenArgs > 0 && flag.Args()[0] == "add_user" {
		if lenArgs != 3 {
			flag.Usage()
			returnCode = FLAGS_FAILURE
			return
		}
		returnCode = FLAGS_ADD_USER
	} else if config.channel == "" {
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

	config, returnCode := getOptions()
	if returnCode == FLAGS_EXIT {
		return
	} else if returnCode == FLAGS_FAILURE {
		log.Fatal("Initialisation failed (getOptions())")
	}
	if config.debug {
		// In debug mode we show the file name and the line from where the log come from
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create the database
	db := database.NewDatabase("", false)
	defer db.Close()

	// Process commands if necessary
	if returnCode == FLAGS_ADD_USER {
		if err := database.AddUser(flag.Args()[1], flag.Args()[2]); err == nil {
			fmt.Println("User added to the database")
		} else {
			fmt.Printf("\nadd_user error: %s\n", err)
		}
		return
	}

	// Create the bot
	bot := core.NewBot(config.nick, config.server, config.channel, config.channelKey)

	// Initialise packages
	for _, module := range config.modules {
		switch strings.TrimSpace(module) {
		case "invoke":
			if !invoke.Init(db, config.emailSender, config.emailAccount, config.emailPassword, config.emailServer, config.emailPort) {
				log.Println("Error while initialising invoke package")
				return
			}
			bot.AddCmdHandler(invoke.HandleInvokeCmd, bot.ReplyToNick)
			help.AddMessages(invoke.HELP_INVOKE)
			log.Println("invoke module loaded")

		case "memo":
			memo.Init(db)
			bot.AddMsgHandler(memo.SendMemo, bot.ReplyToNick)
			bot.AddCmdHandler(memo.HandleMemoCmd, bot.ReplyToAll)
			bot.AddCmdHandler(memo.HandleMemoStatusCmd, bot.ReplyToNick)
			help.AddMessages(memo.HELP_MEMO, memo.HELP_MEMOSTAT)
			log.Println("memo module loaded")

		case "search":
			bot.AddCmdHandler(search.HandleSearchCmd, bot.Reply)
			help.AddMessages(
				search.HELP_DUCKDUCKGO,
				search.HELP_WIKIPEDIA,
				search.HELP_WIKIPEDIA_FR,
				search.HELP_URBANDICTIONNARY)
			log.Println("search module loaded")

		case "webinfo":
			webinfo.Init(db)
			bot.AddMsgHandler(webinfo.HandleUrls, bot.ReplyToAll)
			log.Println("webinfo module loaded")

		case "xkcd":
			bot.AddCmdHandler(xkcd.HandleXKCDCmd, bot.ReplyToAll)
			help.AddMessages(xkcd.HELP_XKCD, xkcd.HELP_XKCD_NUM)
			log.Println("xkcd module loaded")

		default:
		}
	}
	bot.AddCmdHandler(help.HandleHelpCmd, bot.ReplyToNick)

	log.Println("Goxxx started")

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
		log.Printf("System signal received: %s\n", sig)
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
