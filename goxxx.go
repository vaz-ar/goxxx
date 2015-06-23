// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.
package main

import (
	"flag"
	"fmt"
	"github.com/thoj/go-ircevent"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
	"regexp"
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

func getTitleFromHTML(document *html.Node) (title string, found bool) {
	if document.Type != html.DocumentNode {
		// Didn't find a document node as first node, exit
		return
	}

	// Try to find the <html> inside the document
	child := document.FirstChild
	for child != nil && !(child.Type == html.ElementNode && child.Data == "html") {
		child = child.NextSibling
	}
	if child == nil {
		// Didn't find <html>, exit
		return
	}

	// Try to find the <head> inside the document
	currentNode := child
	for child = currentNode.FirstChild;
	    child != nil && !(child.Type == html.ElementNode && child.Data == "head");
	    child = child.NextSibling {
	}
	if child == nil {
		// Didn't find <head>, exit
		return
	}

	// Try to find the <title> inside the <head>
	currentNode = child
	for child = currentNode.FirstChild;
	    child != nil && !(child.Type == html.ElementNode && child.Data == "title");
	    child = child.NextSibling {
	}
	if child == nil || child.FirstChild == nil {
		// Didn't find <title> or it is empty, exit
		return
	}

	// Retrieve the content inside the <title>
	title = child.FirstChild.Data
	found = true

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
		if url.Scheme == "" {
			url.Scheme = "https"
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
	ircConn.Join(channel + " " + channelKey)
	ircConn.AddCallback("PRIVMSG", func(event *irc.Event) {
		allUrls := findUrls(event.Message())
		for _, url := range allUrls {
			fmt.Println("Detected URL:", url.String())
			response, err := http.Get(url.String())
			if err != nil {
				// TODO Do proper logging
				fmt.Println(err)
				return
			}
			doc, err := html.Parse(response.Body)
			response.Body.Close()
			if err != nil {
				// TODO Do proper logging
				fmt.Println(err)
				return
			}
			title, found := getTitleFromHTML(doc)
			if found {
				ircConn.Privmsg(channel, title)
			}
		}
	})

	ircConn.Loop()
}
