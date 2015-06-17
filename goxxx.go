// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.
package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
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

func findUrls(message string) (urls []*url.URL) {
	const maxUrlsCount int = 10

	// Source of the regular expression:
	// http://daringfireball.net/2010/07/improved_regex_for_matching_urls
	re := regexp.MustCompile("(?:https?://|www\\d{0,3}[.]|[a-z0-9.\\-]+[.][a-z]{2,4}/)(?:[^\\s()<>]+|\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\))+(?:\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\)|[^\\s`!()\\[\\]{};:'\".,<>?«»“”‘’])")
	urlCandidates := re.FindAllString(message, maxUrlsCount)

	for _, candidate := range urlCandidates {
		url, err := url.Parse(candidate)
		if err != nil {
			break
		}
		urls = append(urls, url)
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
	ircConn.AddCallback("PRIVMSG", func(event *irc.Event) {
		allUrls := findUrls(event.Message())
		for _, url := range allUrls {
			// TODO Query the URL to retrieve associated content
			fmt.Println(url)
		}
	})

	ircConn.Loop()
}
