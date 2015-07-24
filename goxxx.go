// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/romainletendart/goxxx/core"
	"github.com/romainletendart/goxxx/memo"
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

func initDatabase() *sql.DB {
	// check if the storage directory exist, if not create it
	storage, err := os.Stat("./storage")
	if err != nil {
		os.Mkdir("./storage", os.ModeDir)
	} else if !storage.IsDir() {
		// check if the storage is indeed a directory or not
		log.Fatal("\"storage\" exist but is not a directory")
	}

	db, err := sql.Open("sqlite3", "./storage/db.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {
	nick, server, channel, channelKey, success := getOptions()
	if !success {
		log.Fatal("Initialisation failed (getOptions())")
		return
	}

	database := initDatabase()
	defer database.Close()

	bot := core.Bot{
		Nick:       nick,
		Server:     server,
		Channel:    channel,
		ChannelKey: channelKey,
	}
	bot.Init()
	memo.Init(database)

	bot.AddMsgHandler(webinfo.HandleUrls, bot.ReplyToAll)
	bot.AddCmdHandler(memo.HandleMemoCmd, bot.ReplyToAll)
	bot.AddCmdHandler(memo.SendMemo, bot.ReplyToAll)
	bot.Run()
}
