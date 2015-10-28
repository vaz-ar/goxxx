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
	maxMessages = 5
	sqlInsert   = "INSERT INTO Quote (user, content) VALUES ($1, $2)"
	sqlSelect   = "SELECT content, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')) FROM Quote where user = $1"
	sqlDelete   = "DELETE FROM Quote where user = $1 and content LIKE $2"
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
		HelpMessage: "\t!q/!quote <nick> [<part of message>] (If <part of message> is not supplied, it will list the quotes for <nick>)",
		Triggers:    []string{"!q", "!quote"},
		Handler:     handleQuoteCmd}
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
	} else if len(fields) == 2 {
		return getQuotes(fields, event, callback)
	} else {
		return addQuote(fields, event, callback)
	}
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
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("Quote %q removed for user %q", quote, user)})
	}
	return true
}

func addQuote(fields []string, event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	nick := fields[1]
	size := len(lastMessages[nick])
	max := maxMessages

	if size == 0 {
		return true
	} else if size < max {
		max = size
	}
	var (
		err     error
		matched bool
		msg     string
	)
	for i := max; i >= 1; {
		i--
		msg = lastMessages[nick][i]
		if matched, err = regexp.MatchString(fmt.Sprintf(reMsg, strings.Join(fields[2:], " ")), msg); err != nil {
			log.Fatalf("Quote: Error while matching string (%s)\n", err)
		} else if !matched {
			continue
		}
		_, err := dbPtr.Exec(sqlInsert, nick, msg)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlInsert)
		}
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("Quote %q added for nick %q", msg, nick), Nick: event.Nick})
		break
	}
	return true
}

func getQuotes(fields []string, event *irc.Event, callback func(*core.ReplyCallbackData)) bool {

	rows, err := dbPtr.Query(sqlSelect, fields[1])
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlSelect)
	}
	var content, date string
	for rows.Next() {
		rows.Scan(&content, &date)
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("%q [%s, %s]", content, fields[1], date), Nick: event.Nick})
	}

	return true
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
