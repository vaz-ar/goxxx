// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Package to invoke an user via email
package invoke

import (
	"database/sql"
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/thoj/go-ircevent"
	"log"
	"net/smtp"
	"strings"
)

const (
	HELP_INVOKE = "\t!invoke <nick> [<message>] \t=> Send an email to an user, with an optionnal message " // Help message for the invoke command
	MIN_DELTA   = 15
)

var (
	connection struct {
		auth    smtp.Auth
		account string
		sender  string
		server  string
	}
	dbPtr       *sql.DB // Database pointer
	initialised bool
)

func Init(db *sql.DB, sender, account, password, server string, port int) bool {
	if account == "" || password == "" || server == "" || port == 0 {
		return false
	}
	if sender == "" {
		sender = account
	}
	dbPtr = db

	sqlStmt := `CREATE TABLE IF NOT EXISTS Invoke (
    nick TEXT NOT NULL PRIMARY KEY,
    date DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	connection.account = account
	connection.sender = sender
	connection.server = fmt.Sprint(server, ":", port)
	connection.auth = smtp.PlainAuth(
		"",
		account,
		password,
		server)

	initialised = true
	return true
}

func sendMail(headers map[string]string, body string) bool {
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	err := smtp.SendMail(
		connection.server,
		connection.auth,
		connection.account,
		[]string{headers["To"]},
		[]byte(message))

	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func HandleInvokeCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	if !initialised {
		log.Println("Invoke package not initialised correctly")
		return false
	}
	if callback == nil {
		log.Println("Callback nil for the HandleInvokeCmd function")
		return false
	}

	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1]  => User
	// fields[2]  => Optionnal message (NOT IMPLEMENTED)

	if len(fields) == 0 || fields[0] != "!invoke" {
		return false
	}
	log.Println("Invoke command detected")
	recipient := fields[1]

	sqlQuery := "SELECT ((strftime('%s', datetime('now', 'localtime')) - strftime('%s', date))/60) as delta FROM Invoke WHERE nick = $1"
	var delta int
	err := dbPtr.QueryRow(sqlQuery, recipient).Scan(&delta)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No line for %q in the Invoke table", recipient)
	case err != nil:
		log.Fatalf("%q: %s\n", err, sqlQuery)
	default:
		if delta < MIN_DELTA {
			message := fmt.Sprintf("The user %q was already invoked less than %d minutes ago", recipient, MIN_DELTA)
			log.Println(message)
			callback(&core.ReplyCallbackData{
				Message: message,
				Nick:    event.Nick})
			return true
		}
	}

	sqlQuery = "SELECT email FROM User WHERE nick = $1"
	var email string
	err = dbPtr.QueryRow(sqlQuery, recipient).Scan(&email)
	switch {
	case err == sql.ErrNoRows:
		message := fmt.Sprintf("No user in the datbase with %q for nick, call the cops! (or maybe just the bot admin)", recipient)
		log.Println(message)
		callback(&core.ReplyCallbackData{
			Message: message,
			Nick:    event.Nick})
	case err != nil:
		log.Fatalf("%q: %s\n", err, sqlQuery)
	default:
	}

	headers := map[string]string{
		"From":    connection.sender,
		"To":      email,
		"Subject": "Goxxx: Your presence is requested on " + event.Arguments[0]}

	if !sendMail(headers, fmt.Sprintf("Your presence has been requested by %s on the %s channel.\n Hurry up!\n", event.Nick, event.Arguments[0])) {
		log.Println("Invoke command: sendMail failed to send the email")
		callback(&core.ReplyCallbackData{
			Message: "The invoke command failed, the email was not sent",
			Nick:    event.Nick})
		return true
	}
	log.Println("Invoke command: email sent")

	sqlQuery = "INSERT OR REPLACE INTO Invoke (nick) VALUES ($1)"
	_, err = dbPtr.Exec(sqlQuery, recipient)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}

	callback(&core.ReplyCallbackData{
		Message: fmt.Sprintf("Email sent to %s", recipient),
		Nick:    event.Nick})

	return true
}
