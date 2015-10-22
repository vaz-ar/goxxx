// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

/*
Package pictures contains picture related commands
*/
package pictures

import (
	"database/sql"
	"fmt"
	"github.com/emirozer/go-helpers"
	"github.com/thoj/go-ircevent"
	"github.com/vaz-ar/goxxx/core"
	"log"
	"path"
	"regexp"
	"strings"
)

const (
	// HelpPictures is the help message for the !pic command
	HelpPictures = "\t!p/!pic <search terms> \t=> Search in the database for pictures matching <search terms>"
	// HelpPicturesAdd is the help message for the !addpic command
	HelpPicturesAdd = "\t!addpic <tag> <url>  \t\t\t=> Add a picture in the database for <tag> (<url> must have an image extension)"
	maxPictures     = 10
)

var (
	dbPtr   *sql.DB // Database pointer
	extList = []string{".png", ".jpg", ".jpeg"}
	// Source of the regular expression:
	// http://daringfireball.net/2010/07/improved_regex_for_matching_urls
	reURL = regexp.MustCompile("(?:https?://|www\\d{0,3}[.]|[a-z0-9.\\-]+[.][a-z]{2,4}/)(?:[^\\s()<>]+|\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\))+(?:\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\)|[^\\s`!()\\[\\]{};:'\".,<>?«»“”‘’])")
)

// Init stores the database pointer and initialises the database table "Pictures" if necessary.
func Init(db *sql.DB) {
	dbPtr = db
	sqlStmt := `CREATE TABLE IF NOT EXISTS Picture (
    id integer NOT NULL PRIMARY KEY,
    tag TEXT,
    url TEXT,
    nick TEXT,
    date DATETIME DEFAULT CURRENT_TIMESTAMP);`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
}

// HandlePictureCmd returns the pictures associated with a tag
func HandlePictureCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	if callback == nil {
		log.Println("Callback nil for the HandlePictureCmd function")
		return false
	}
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => Tag to search for
	if len(fields) < 2 || (fields[0] != "!p" && fields[0] != "!pic") {
		return false
	}

	var tag, url string
	sqlQuery := "SELECT tag, url FROM Picture WHERE tag LIKE $1"
	rows, err := dbPtr.Query(sqlQuery, "%"+strings.Join(fields[1:], " ")+"%")
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	for rows.Next() {
		rows.Scan(&tag, &url)
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("Picture for %q : %s", tag, url)})
	}

	return true
}

// HandleAddPictureCmd add a picture for a given tag to the database
func HandleAddPictureCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	if callback == nil {
		log.Println("Callback nil for the HandleAddPictureCmd function")
		return false
	}
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1] => url for the picture
	// fields[2:] => Tag for the picture
	if len(fields) < 3 || fields[0] != "!addpic" {
		return false
	}
	url := fields[1]
	if !reURL.MatchString(url) || !helpers.StringInSlice(strings.ToLower(path.Ext(url)), extList) {
		callback(&core.ReplyCallbackData{Message: "Incorrect format for the \"Add Picture\" command (see !help)"})
		return true
	}

	var (
		tag      = strings.Join(fields[2:], " ")
		count    int
		sqlQuery = "SELECT count(url) FROM Picture WHERE tag = $1"
	)
	err := dbPtr.QueryRow(sqlQuery, tag).Scan(&count)
	if err != sql.ErrNoRows && err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	if count >= maxPictures {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("There is already too much pictures for the tag %q", tag)})
		return true
	}

	sqlQuery = "SELECT tag FROM Picture WHERE url = $1"
	rows, err := dbPtr.Query(sqlQuery, url)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	if rows.Next() {
		callback(&core.ReplyCallbackData{Message: fmt.Sprintf("This picture is already present for the tag %q", tag)})
		return true
	}

	sqlQuery = "INSERT INTO Picture (tag, url, nick) VALUES ($1, $2, $3)"
	_, err = dbPtr.Exec(sqlQuery, tag, url, event.Nick)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlQuery)
	}
	callback(&core.ReplyCallbackData{Message: fmt.Sprintf("Picture %q added for tag %q", url, tag)})

	return true
}
