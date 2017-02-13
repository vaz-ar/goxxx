// The MIT License (MIT)
//
// Copyright (c) 2015 Romain LÉTENDART
//
// See LICENSE file.

// Package webinfo retrieves informations from links
package webinfo

import (
	"database/sql"
	"fmt"
	"github.com/emirozer/go-helpers"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"golang.org/x/net/html"
	"golang.org/x/net/idna"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	// maxUrlsCount Maximun number of URLs to search in one message
	maxUrlsCount        = 10
	sqlSelectExist      = "SELECT user, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')) FROM Link WHERE url = $1"
	sqlSelectWhereTitle = "SELECT user, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), title, url FROM Link WHERE title LIKE $1"
	sqlSelectWhereUrl   = "SELECT user, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), title, url FROM Link WHERE url LIKE $1"
	sqlInsert           = "INSERT INTO Link (user, url, title) VALUES ($1, $2, $3)"
)

var (
	dbPtr        *sql.DB                                // Database pointer
	urlShortener = []string{"t.co", "bit.ly", "goo.gl"} // URL shorteners base URL
)

// GetCommand returns a Command structure for url search command
func GetTitleCommand() *core.Command {
	return &core.Command{
		Module:      "url",
		HelpMessage: "\t!urlt <search terms>\t\t\t\t\t\t=> Return links with titles matching <search terms>",
		Triggers:    []string{"!urlt"},
		Handler:     handleSearchTitlesCmd}
}
func GetUrlCommand() *core.Command {
	return &core.Command{
		Module:      "url",
		HelpMessage: "\t!url <search terms>\t\t\t\t\t\t=> Return links with urls matching <search terms>",
		Triggers:    []string{"!url"},
		Handler:     handleSearchUrlsCmd}
}

// Init stores the database pointer.
func Init(db *sql.DB) {
	dbPtr = db
}

// BUG(romainletendart) Choose a better name for the HandleUrls function

// HandleURLs is a message handler that search for URLs in a message
func HandleURLs(event *irc.Event, callback func(*core.ReplyCallbackData)) {

	client := &http.Client{}

	for _, currentURL := range findURLs(event.Message()) {

		log.Println("Detected URL:", currentURL.String())

		req, err := http.NewRequest("GET", currentURL.String(), nil)
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Set("User-Agent", "Goxxx/1.0")

		response, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return
		}
		defer response.Body.Close()

		doc, err := html.Parse(response.Body)
		if err != nil {
			log.Println(err)
			return
		}

		var user, date string
		// BUG(vaz-ar) Maybe not necessary to use Query + loop here, see if QueryRow can do the trick
		rows, err := dbPtr.Query(sqlSelectExist, currentURL.String())
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlSelectExist)
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&user, &date)
		}

		if user != "" {
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("Link already posted by %s (%s)", user, date),
				Target:  core.GetTargetFromEvent(event)})
		}

		title, found := getTitleFromHTML(doc)
		if found {
			log.Println("Title found: ", title)
			if helpers.StringInSlice(currentURL.Host, urlShortener) {
				title += fmt.Sprint(" (", response.Request.URL.String(), ")")
			}
			callback(&core.ReplyCallbackData{
				Message: title,
				Target:  core.GetTargetFromEvent(event)})
		} else {
			log.Println("No title found for ", currentURL.String())
		}

		// If the link was not found we save it in the database along with the user that posted it and it's title
		if user == "" {
			_, err := dbPtr.Exec(sqlInsert, event.Nick, currentURL.String(), title)
			if err != nil {
				log.Fatalf("%q: %s\n", err, sqlInsert)
			}
		}
	}
}

// handleSearchTitlesCmd is a command handler that search in the database for page titles matching a pattern
func handleSearchTitlesCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:]  => URL
	if len(fields) < 2 {
		return false
	}
	var (
		user, date, title, url string
		search                 = strings.Join(fields[1:], " ")
	)
	// BUG(vaz-ar) Maybe not necessary to use Query + loop here, see if QueryRow can do the trick
	rows, err := dbPtr.Query(sqlSelectWhereTitle, "%"+search+"%")
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlSelectWhereTitle)
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&user, &date, &title, &url)
		if title == "" {
			title = "No Title"
		}
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf(`Link found for "%s" => %s (%s) [Posted by %s, %s]`, search, title, url, user, date),
			Target:  core.GetTargetFromEvent(event)})
	}
	return true
}

// handleSearchUrlsCmd is a command handler that search in the database for url matching a pattern
func handleSearchUrlsCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:]  => URL
	if len(fields) < 2 {
		return false
	}
	var (
		user, date, title, url string
		search                 = strings.Join(fields[1:], " ")
	)
	// BUG(vaz-ar) Maybe not necessary to use Query + loop here, see if QueryRow can do the trick
	rows, err := dbPtr.Query(sqlSelectWhereUrl, "%"+search+"%")
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlSelectWhereUrl)
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&user, &date, &title, &url)
		if title == "" {
			title = "No Title"
		}
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf(`URLs matching "%s" => %s (%s) [Posted by %s, %s]`, search, title, url, user, date),
			Target:  core.GetTargetFromEvent(event)})
	}
	return true
}

// Extract the title from an HTML page
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
	for child = currentNode.FirstChild; child != nil && !(child.Type == html.ElementNode && child.Data == "head"); child = child.NextSibling {
	}
	if child == nil {
		// Didn't find <head>, exit
		return
	}

	// Try to find the <title> inside the <head>
	currentNode = child
	for child = currentNode.FirstChild; child != nil && !(child.Type == html.ElementNode && child.Data == "title"); child = child.NextSibling {
	}
	if child == nil || child.FirstChild == nil {
		// Didn't find <title> or it is empty, exit
		return
	}

	// Retrieve the content inside the <title> and post-process it
	title = strings.TrimSpace(child.FirstChild.Data)
	// Replace every whitespaces or new lines sequence with a single
	// whitespace
	re := regexp.MustCompile("\\s+")
	title = re.ReplaceAllString(title, " ")
	found = true

	return
}

// Search for URLs in a string
func findURLs(message string) (urls []*url.URL) {
	// Source of the regular expression:
	// http://daringfireball.net/2010/07/improved_regex_for_matching_urls
	re := regexp.MustCompile("(?:https?://|www\\d{0,3}[.]|[a-z0-9.\\-]+[.][a-z]{2,4}/)(?:[^\\s()<>]+|\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\))+(?:\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\)|[^\\s`!()\\[\\]{};:'\".,<>?«»“”‘’])")
	urlCandidates := re.FindAllString(message, maxUrlsCount)

	for _, candidate := range urlCandidates {
		url, err := url.Parse(candidate)
		if err != nil {
			break
		}
		// Scheme is required to query a URL
		if url.Scheme == "" {
			url.Scheme = "http"
		}
		// Conversion to ASCII is needed for Unicode hostnames
		asciiHost, err := idna.ToASCII(url.Host)
		if err == nil {
			url.Host = asciiHost
		}
		urls = append(urls, url)
	}
	return
}
