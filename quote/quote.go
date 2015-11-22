// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

/*
Package quote contains quote commands
*/
package quote

import (
	"database/sql"
	"fmt"
	"github.com/emirozer/go-helpers"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"log"
	"regexp"
	"strings"
)

const (
	maxMessages  = 10
	sqlInsert    = "INSERT INTO Quote (user, content, sender) VALUES ($1, $2, $3)"
	sqlSelect    = "SELECT content, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), sender FROM Quote WHERE user = $1 AND content LIKE $2"
	sqlSelectAll = "SELECT content, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), sender FROM Quote WHERE user = $1"
	sqlDelete    = "DELETE FROM Quote where user = $1 AND content LIKE $2"
)

var (
	dbPtr          *sql.DB // Database pointer
	lastMessages   map[string][]string
	reMsg          = `.*%s.*`
	administrators []string
)

// GetQuoteCommand returns a Command structure for the quote command
func GetQuoteCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "\t!q/!quote <nick> [<part of message>]",
		Triggers:    []string{"!q", "!quote"},
		Handler:     handleQuoteCmd}
}

// GetAddQuoteCommand returns a Command structure for the addquote command
func GetAddQuoteCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "\t!aq/!addquote <nick> <part of message>",
		Triggers:    []string{"!aq", "!addquote"},
		Handler:     handleAddQuoteCmd}
}

// GetRmQuoteCommand returns a Command structure for the remove quote command
func GetRmQuoteCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "\t!rmq/!rmquote <nick> <part of the quote> (Admins only)",
		Triggers:    []string{"!rmq", "!rmquote"},
		Handler:     handleRmQuoteCmd}
}

// Init stores the database pointer and initialises the database table "Quote" if necessary.
func Init(db *sql.DB, admins []string) {
	dbPtr = db
	sqlStmt := `CREATE TABLE IF NOT EXISTS Quote (
    id integer NOT NULL PRIMARY KEY,
    user TEXT,
    content TEXT,
    date DATETIME DEFAULT CURRENT_TIMESTAMP);`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	lastMessages = make(map[string][]string)
	administrators = admins
}

// handleQuoteCmd
func handleQuoteCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1] => Nick
	// fields[2:] => part of the message to search for
	if len(fields) == 1 {
		return true
	}
	var (
		rows *sql.Rows
		err  error
	)
	if len(fields) >= 3 {
		// Search with part of the message
		messagePart := prepareForSearch(strings.Join(fields[2:], " "))
		rows, err = dbPtr.Query(sqlSelect, fields[1], "%"+messagePart+"%")
	} else {
		// Search without part of the message
		rows, err = dbPtr.Query(sqlSelectAll, fields[1])
	}
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlSelect)
	}
	defer rows.Close()

	var content, date, sender string
	for rows.Next() {
		rows.Scan(&content, &date, &sender)
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("%q [%s, %s, quoted by %s]", content, fields[1], date, sender),
			Nick:    event.Nick})
	}

	return true
}

func handleAddQuoteCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())

	if len(fields) < 3 {
		return false
	}

	nick := fields[1]
	size := len(lastMessages[nick])
	max := maxMessages

	if size == 0 {
		return true
	} else if size < max {
		max = size
	}

	var (
		rawMsg   string
		cleanMsg string
		pattern  = prepareForSearch(strings.Join(fields[2:], " "))
	)
	// Look for the search pattern in one of the last messages from "nick"
	for i := max; i >= 1; {
		i--
		rawMsg = lastMessages[nick][i]
		cleanMsg = prepareForSearch(rawMsg)
		if !strings.Contains(cleanMsg, pattern) {
			continue
		}
		_, err := dbPtr.Exec(sqlInsert, nick, rawMsg, event.Nick)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlInsert)
		}
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("Quote %q added for nick %q", rawMsg, nick), Nick: event.Nick})
		break
	}
	return true
}

// handleRmQuoteCmd
func handleRmQuoteCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1] => Nick
	// fields[2:] => part of the quote to search for
	if len(fields) < 3 {
		return false
	}

	if !helpers.StringInSlice(event.Nick, administrators) {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("You need to be an administrator to run this command (Admins: %q)", strings.Join(administrators, ", "))})
		return true
	}

	quote := strings.Join(fields[2:], " ")
	user := fields[1]
	result, err := dbPtr.Exec(sqlDelete, user, "%"+quote+"%")
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlDelete)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlDelete)
	}
	if rows != 0 {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("Quote(s) matching %%%q%% removed for user %q", quote, user)})
	}
	return true
}

func prepareForSearch(message string) string {
	// Replace basic punctuation with a single space
	re := regexp.MustCompile(`[.,;:!'"-]`)
	message = re.ReplaceAllString(message, " ")

	// Replace every whitespaces or new lines sequence with a single space
	re = regexp.MustCompile("\\s+")
	message = re.ReplaceAllString(message, " ")

	// Remove case-sentivity using only lowercase
	return strings.TrimSpace(strings.ToLower(message))
}

// HandleMessages is a message handler that stores the last messages by users
func HandleMessages(event *irc.Event, callback func(*core.ReplyCallbackData)) {
	if len(lastMessages[event.Nick]) < maxMessages {
		lastMessages[event.Nick] = append(lastMessages[event.Nick], event.Message())
	} else {
		lastMessages[event.Nick] = append(lastMessages[event.Nick][:0], lastMessages[event.Nick][1:]...)
		lastMessages[event.Nick] = append(lastMessages[event.Nick], event.Message())
	}
}
