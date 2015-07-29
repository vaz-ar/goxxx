// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.
package memo

import (
	"database/sql"
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/thoj/go-ircevent"
	"log"
	"strings"
)

var _database *sql.DB

type memoData struct {
	id        int
	date      string
	message   string
	user_from string
	user_to   string
}

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

func HandleMemoCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1]  => recipient's nick
	// fields[2:] => message
	if len(fields) == 0 || fields[0] != "!memo" {
		return false
	}

	if len(fields) < 3 {
		if callback != nil {
			callback(&core.ReplyCallbackData{
				Message: fmt.Sprintf("Memo usage: %s \"recipient's nick\" \"message\"", fields[0])})
			callback(&core.ReplyCallbackData{
				Message: "Memo usage: !memostat to list unread memos"})
		}
		return false
	}

	memo := memoData{
		user_to:   fields[1],
		user_from: event.Nick,
		message:   strings.Join(fields[2:], " ")}

	sqlStmt := "INSERT INTO Memo (user_to, user_from, message) VALUES ($1, $2, $3)"
	_, err := _database.Exec(sqlStmt, memo.user_to, memo.user_from, memo.message)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	if callback != nil {
		callback(&core.ReplyCallbackData{
			// Message: fmt.Sprintf("Memo for %s from %s: \"%s\"", fields[1], user_from, message)})
			Message: fmt.Sprintf("%s: memo for %s saved", memo.user_from, memo.user_from)})
	}
	return true
}

func SendMemo(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
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

	var memoList []memoData
	for rows.Next() {
		var memo memoData
		rows.Scan(&memo.id, &memo.user_to, &memo.user_from, &memo.message, &memo.date)
		memoList = append(memoList, memo)
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("%s: memo from %s => \"%s\" (%s)", memo.user_to, memo.user_from, memo.message, memo.date)})
	}
	rows.Close()

	for _, memo := range memoList {
		sqlQuery = "DELETE FROM Memo WHERE id = $1"
		_, err = _database.Exec(sqlQuery, memo.id)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlQuery)
		}
	}
	return false
}

func HandleMemoStatusCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	if strings.TrimSpace(event.Message()) != "!memostat" {
		return false
	}

	sqlQuery := "SELECT id, user_to, message, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')) FROM Memo WHERE user_from = $1 ORDER BY id"
	rows, err := _database.Query(sqlQuery, event.Nick)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	defer rows.Close()

	var memo memoData
	for rows.Next() {
		rows.Scan(&memo.id, &memo.user_to, &memo.message, &memo.date)
		callback(&core.ReplyCallbackData{fmt.Sprintf("Memo for %s: %q (%s)", memo.user_to, memo.message, memo.date), event.Nick})
	}
	rows.Close()

	if memo.id == 0 {
		callback(&core.ReplyCallbackData{"No memo saved", event.Nick})
	}

	return true
}
