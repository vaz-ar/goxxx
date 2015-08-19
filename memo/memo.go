// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Memo package
package memo

import (
	"database/sql"
	"fmt"
	"github.com/romainletendart/goxxx/core"
	"github.com/thoj/go-ircevent"
	"log"
	"strings"
)

const (
	HELP_MEMO     string = "\t!memo/!m <nick> <message> \t=> Leave a memo for another user"                               // Help message for the memo command
	HELP_MEMOSTAT string = "\t!memostat/!ms \t\t\t\t\t=> Get the list of the unread memos (List only the memos you left)" // Help message for the memo status command
)

var (
	memo_cmd     = []string{"!memo", "!m"}      // Slice containing the possible memo commands
	memostat_cmd = []string{"!memostat", "!ms"} // Slice containing the possible memo status commands
	dbPtr        *sql.DB                        // Database pointer
)

// Used to store memo data, based on the database table "Memo".
type MemoData struct {
	id        int
	Date      string
	Message   string
	User_from string
	User_to   string
}

// Store the database pointer and initialise the database table "Memo" if necessary.
func Init(db *sql.DB) {
	dbPtr = db
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

// Handler for the memo command.
func HandleMemoCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1]  => recipient's nick
	// fields[2:] => message
	if len(fields) < 3 || !core.StringInSlice(fields[0], memo_cmd) {
		return false
	}
	memo := MemoData{
		User_to:   fields[1],
		User_from: event.Nick,
		Message:   strings.Join(fields[2:], " ")}

	sqlStmt := "INSERT INTO Memo (user_to, user_from, message) VALUES ($1, $2, $3)"
	_, err := dbPtr.Exec(sqlStmt, memo.User_to, memo.User_from, memo.Message)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	if callback != nil {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("%s: memo for %s saved", memo.User_from, memo.User_to),
			Nick:    memo.User_from})
	}
	return true
}

// Message handler, will send memo(s) to an user when he post a message for the first time after a memo for him was created.
func SendMemo(event *irc.Event, callback func(*core.ReplyCallbackData)) {
	user := event.Nick
	sqlQuery := "SELECT id, user_from, message, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')) FROM Memo WHERE user_to = $1;"
	rows, err := dbPtr.Query(sqlQuery, user)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	defer rows.Close()

	user_to := event.Nick
	var memoList []MemoData
	for rows.Next() {
		var memo MemoData
		rows.Scan(&memo.id, &memo.User_from, &memo.Message, &memo.Date)
		memoList = append(memoList, memo)
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("%s: memo from %s => \"%s\" (%s)", user_to, memo.User_from, memo.Message, memo.Date),
			Nick:    user_to})
	}
	rows.Close()

	for _, memo := range memoList {
		sqlQuery = "DELETE FROM Memo WHERE id = $1"
		_, err = dbPtr.Exec(sqlQuery, memo.id)
		if err != nil {
			log.Fatalf("%q: %s\n", err, sqlQuery)
		}
	}
}

// Handler for the memo status command.
func HandleMemoStatusCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	if len(fields) == 0 || !core.StringInSlice(fields[0], memostat_cmd) {
		return false
	}

	sqlQuery := "SELECT id, user_to, message, strftime('%d/%m/%Y @ %H:%M', datetime(date, 'localtime')) FROM Memo WHERE user_from = $1 ORDER BY id"
	rows, err := dbPtr.Query(sqlQuery, event.Nick)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	defer rows.Close()

	var memo MemoData
	for rows.Next() {
		rows.Scan(&memo.id, &memo.User_to, &memo.Message, &memo.Date)
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Memo for %s: \"%s\" (%s)", memo.User_to, memo.Message, memo.Date),
			Nick:    event.Nick})
	}
	rows.Close()

	if memo.id == 0 {
		callback(&core.ReplyCallbackData{Message: "No memo saved", Nick: event.Nick})
	}
	return true
}
