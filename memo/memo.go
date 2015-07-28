// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package memo

import (
	"database/sql"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"strings"
)

var _database *sql.DB

func Init(db *sql.DB) {
	_database = db
	sqlStmt := `CREATE TABLE IF NOT EXISTS Memo (
    id integer NOT NULL PRIMARY KEY,
    user_to TEXT,
    user_from TEXT,
    message TEXT,
    date DATETIME DEFAULT CURRENT_TIMESTAMP);`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
}

func HandleMemoCmd(event *irc.Event, callback func(string)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1]  => recipient's nick
	// fields[2:] => message
	if len(fields) == 0 || fields[0] != "!memo" {
		return false
	}

	if len(fields) < 3 {
		if callback != nil {
			callback(fmt.Sprintf("Memo usage: %s \"recipient's nick\" \"message\"", fields[0]))
		}
		return false
	}
	user_from := event.Nick
	message := strings.Join(fields[2:], " ")

	sqlStmt := "INSERT INTO Memo (user_to, user_from, message) VALUES ($1, $2, $3)"
	_, err := _database.Exec(sqlStmt, fields[1], user_from, message)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	if callback != nil {
		callback(fmt.Sprintf("Memo for %s from %s: \"%s\"", fields[1], user_from, message))
	}
	return true

}

func SendMemo(event *irc.Event, callback func(string)) bool {
	if callback == nil {
		log.Println("Callback nil for the SendMemo function, unable to send the memo")
	}

	user := event.Nick
	sqlQuery := "SELECT id, user_to, user_from, message, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')) FROM Memo WHERE user_to = $1;"
	rows, err := _database.Query(sqlQuery, user)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	defer rows.Close()

	var (
		id        int
		date      string
		message   string
		user_from string
		user_to   string
		idList    []int
	)

	for rows.Next() {
		rows.Scan(&id, &user_to, &user_from, &message, &date)
		idList = append(idList, id)
		callback(fmt.Sprintf("%s: %s left you a memo => \"%s\" (%s)", user_to, user_from, message, date))
	}
	rows.Close()

	for _, id = range idList {
		sqlQuery = "DELETE FROM Memo WHERE id = $1"
		_, err = _database.Exec(sqlQuery, id)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlQuery)
		}
	}
	return false
}
