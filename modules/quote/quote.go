// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
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
	maxMessages           = 20
	sqlInsert             = "INSERT INTO Quote (user, content, sender) VALUES ($1, $2, $3)"
	sqlSelect             = "SELECT content, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), sender FROM Quote WHERE user = $1 AND content LIKE $2"
	sqlSelectFromAll      = "SELECT content, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), sender, user FROM Quote WHERE content LIKE $1"
	sqlSelectExactContent = "SELECT sender FROM Quote WHERE user = $1 AND content = $2"
	sqlSelectAll          = "SELECT content, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), sender FROM Quote WHERE user = $1"
	sqlSelectFromDay      = "SELECT content, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')), sender, user FROM Quote WHERE date(date, 'localtime') = date('now', '-1 year', 'localtime') ORDER BY RANDOM() LIMIT 1"
	sqlDelete             = "DELETE FROM Quote where user = $1 AND content LIKE $2"
)

var (
	dbPtr          *sql.DB // Database pointer
	lastMessages   map[string][]string
	reMsg          = `.*%s.*`
	administrators *[]string
)

// GetQuoteCommand returns a Command structure for the quote command
func GetQuoteCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "!q/!quote <nick> [<part of message>]",
		Triggers:    []string{"!q", "!quote"},
		Handler:     handleQuoteCmd}
}

// GetQuoteFromAllCommand returns a Command structure for the quote all command
func GetQuoteFromAllCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "!qa/!quoteall [<part of message>]",
		Triggers:    []string{"!qa", "!quoteall"},
		Handler:     handleQuoteAllCmd}
}

// GetAddQuoteCommand returns a Command structure for the addquote command
func GetAddQuoteCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "!aq/!addquote <nick> <part of message>",
		Triggers:    []string{"!aq", "!addquote"},
		Handler:     handleAddQuoteCmd}
}

// GetRmQuoteCommand returns a Command structure for the remove quote command
func GetRmQuoteCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "!rmq/!rmquote <nick> <part of the quote> (Admins only)",
		Triggers:    []string{"!rmq", "!rmquote"},
		Handler:     handleRmQuoteCmd}
}

// GetDailyQuoteCommand returns a Command structure for the daily quote command
func GetDailyQuoteCommand() *core.Command {
	return &core.Command{
		Module:      "quote",
		HelpMessage: "!dq (No parameter needed)",
		Triggers:    []string{"!dq"},
		Handler:     handleDailyQuoteCmd}
}

// Init stores the database pointer and the administrators list.
func Init(db *sql.DB, admins *[]string) {
	lastMessages = make(map[string][]string)
	dbPtr = db
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
		log.Fatalf("\"%s\": %s\n", err, sqlSelect)
	}
	defer rows.Close()

	var content, date, sender string
	for rows.Next() {
		rows.Scan(&content, &date, &sender)
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("%s [%s, %s, quoted by %s]", content, fields[1], date, sender),
			Target:  core.GetTargetFromEvent(event)})
	}

	return true
}

// handleQuoteAllCmd
func handleQuoteAllCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => part of the message to search for
	if len(fields) == 1 {
		return true
	}

	// Search with part of the message
	messagePart := prepareForSearch(strings.Join(fields[1:], " "))
	rows, err := dbPtr.Query(sqlSelectFromAll, "%"+messagePart+"%")

	if err != nil {
		log.Fatalf("\"%s\": %s\n", err, sqlSelectFromAll)
	}
	defer rows.Close()

	var content, date, sender, user string
	for rows.Next() {
		rows.Scan(&content, &date, &sender, &user)
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("%s [%s, %s, quoted by %s]", content, user, date, sender),
			Target:  core.GetTargetFromEvent(event)})
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

		// Check if quote already exists in the database
		rows, err := dbPtr.Query(sqlSelectExactContent, nick, rawMsg)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlSelectExactContent)
		}
		defer rows.Close()
		if rows.Next() {
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("This quote is already present for the user \"%s\"", nick),
				Target:  core.GetTargetFromEvent(event)})
			return true
		}

		// Insert quote in the database
		_, err = dbPtr.Exec(sqlInsert, nick, rawMsg, event.Nick)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlInsert)
		}
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Quote \"%s\" added for nick \"%s\"", rawMsg, nick),
			Target:  core.GetTargetFromEvent(event)})
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

	// update the administrators list
	core.UpdateUserList(event)

	if !helpers.StringInSlice(event.Nick, *administrators) {
		if len(*administrators) > 1 {
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("You need to be an administrator to run this command (Admins: \"%s\")", strings.Join(*administrators, ", ")),
				Target:  core.GetTargetFromEvent(event)})
		} else if len(*administrators) == 1 {
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("You need to be an administrator to run this command (Admin: \"%s\")", (*administrators)[0]),
				Target:  core.GetTargetFromEvent(event)})
		} else {
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintln("You need to be an administrator to run this command (No admin set!)"),
				Target:  core.GetTargetFromEvent(event)})
		}
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
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Quote(s) matching \"%%%s%%\" removed for user \"%s\"", quote, user),
			Target:  core.GetTargetFromEvent(event)})
	}
	return true
}

// handleDailyQuoteCmd
func handleDailyQuoteCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {

	rows, err := dbPtr.Query(sqlSelectFromDay)

	if err != nil {
		log.Fatalf("\"%s\": %s\n", err, sqlSelectFromDay)
	}
	defer rows.Close()

	if rows.Next() {
		var content, date, sender, user string
		rows.Scan(&content, &date, &sender, &user)
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("%s [%s, %s, quoted by %s]", content, user, date, sender),
			Target:  core.GetTargetFromEvent(event)})
		return true
	}

	callback(&core.ReplyCallbackData{
		Message: "There was no quote 1 year ago, losers!",
		Target:  core.GetTargetFromEvent(event)})

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
