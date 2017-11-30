// The MIT License (MIT)
//
// Copyright (c) 2017 Arnaud Vazard
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
	maxPictures           = 5
	sqlInsert             = "INSERT INTO Picture (tag, url, nick, nsfw) VALUES ($1, $2, $3, $4)"
	sqlSelectTagWhereURL  = "SELECT tag FROM Picture WHERE url = $1 AND tag = $2"
	sqlCount              = "SELECT count(url) FROM Picture WHERE tag = $1"
	sqlDelete             = "DELETE FROM Picture WHERE tag = $1 AND url = $2"
	sqlSelectWhereTagLike = "SELECT tag, url, nsfw FROM Picture WHERE tag LIKE $1"
	sqlSelectAll          = "SELECT tag, url, nsfw FROM Picture ORDER BY tag"
)

var (
	dbPtr          *sql.DB // Database pointer
	extList        = []string{".png", ".jpg", ".jpeg"}
	administrators *[]string
	// Source of the regular expression: http://daringfireball.net/2010/07/improved_regex_for_matching_urls
	reURL      = regexp.MustCompile("(?:https?://|www\\d{0,3}[.]|[a-z0-9.\\-]+[.][a-z]{2,4}/)(?:[^\\s()<>]+|\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\))+(?:\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\)|[^\\s`!()\\[\\]{};:'\".,<>?«»“”‘’])")
	reSanitize = regexp.MustCompile(`[%?_$:@]`)
)

// GetPicCommand returns a Command structure for the picture command
func GetPicCommand() *core.Command {
	return &core.Command{
		Module:      "pictures",
		HelpMessage: "!p/!pic <search terms> => Search in the database for pictures matching <search terms>",
		Triggers:    []string{"!p", "!pic"},
		Handler:     handlePictureCmd}
}

// GetAddPicCommand returns a Command structure for the add picture command
func GetAddPicCommand() *core.Command {
	return &core.Command{
		Module:      "pictures",
		HelpMessage: "!ap/!addpic <url> <tag> [#NSFW] => Add a picture in the database for <tag> (<url> must have an image extension)",
		Triggers:    []string{"!ap", "!addpic"},
		Handler:     handleAddPictureCmd}
}

// GetRmPicCommand returns a Command structure for the remove picture command
func GetRmPicCommand() *core.Command {
	return &core.Command{
		Module:      "pictures",
		HelpMessage: "!rmpic <url> <tag> => Remove a picture in the database for <tag> (Admin only command)",
		Triggers:    []string{"!rmpic"},
		Handler:     handleRmPictureCmd}
}

// Init stores the database pointer and the administrators list.
func Init(db *sql.DB, admins *[]string) {
	dbPtr = db
	administrators = admins
}

// handlePictureCmd returns the pictures associated with a tag
func handlePictureCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1:] => Tag to search for
	if len(fields) < 2 {
		return false
	}

	var (
		requestedTag = prepareTagString(strings.Join(fields[1:], " "))
		rows         *sql.Rows
		err          error
	)
	if requestedTag == "" {
		callback(&core.ReplyCallbackData{
			Message: "Picture command: No data remaining for the tag value after sanitization.",
			Target:  core.GetTargetFromEvent(event)})
		return true
	}

	rows, err = dbPtr.Query(sqlSelectWhereTagLike, "%"+requestedTag+"%")
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlSelectWhereTagLike)
	}

	defer rows.Close()

	var (
		message, tag, url string
		resultCount, nsfw int
	)
	for rows.Next() {
		resultCount++
		rows.Scan(&tag, &url, &nsfw)
		if nsfw == 0 {
			message = "Picture for \"%s\" : %s"
		} else {
			message = "Picture for \"%s\" (#NSFW) : %s"
		}
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf(message, tag, url),
			Target:  core.GetTargetFromEvent(event)})
	}
	if resultCount == 0 {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("No picture found for tag \"%s\"", requestedTag),
			Target:  core.GetTargetFromEvent(event)})
	}

	return true
}

// handleAddPictureCmd add a picture for a given tag to the database
func handleAddPictureCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1] => url for the picture
	// fields[2:] => Tag for the picture
	if len(fields) < 3 {
		return false
	}
	url := fields[1]
	if !reURL.MatchString(url) || !helpers.StringInSlice(strings.ToLower(path.Ext(url)), extList) {
		callback(&core.ReplyCallbackData{
			Message: "Incorrect format for the \"Add Picture\" command (see !help)",
			Target:  core.GetTargetFromEvent(event)})
		return true
	}

	var (
		tag   = prepareTagString(strings.Join(fields[2:], " "))
		nsfw  = prepareTagString(fields[len(fields)-1]) == "#nsfw"
		count int
	)
	// Check if last element from fields is NSFW tag
	if nsfw {
		tag = strings.TrimSpace(strings.TrimSuffix(tag, "#nsfw"))
	}
	err := dbPtr.QueryRow(sqlCount, tag).Scan(&count)
	if err != sql.ErrNoRows && err != nil {
		log.Fatalf("%q: %s\n", err, sqlCount)
	}
	if count >= maxPictures {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("There is already too much pictures for the tag \"%s\"", tag),
			Target:  core.GetTargetFromEvent(event)})
		return true
	}

	rows, err := dbPtr.Query(sqlSelectTagWhereURL, url, tag)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlSelectTagWhereURL)
	}
	defer rows.Close()

	if rows.Next() {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("This picture is already present for the tag \"%s\"", tag),
			Target:  core.GetTargetFromEvent(event)})
		return true
	}

	_, err = dbPtr.Exec(sqlInsert, tag, url, event.Nick, nsfw)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlInsert)
	}
	callback(&core.ReplyCallbackData{
		Message: fmt.Sprintf("Picture \"%s\" added for tag \"%s\"", url, tag),
		Target:  core.GetTargetFromEvent(event)})

	return true
}

// handleRmPictureCmd remove a picture for a given tag to the database
func handleRmPictureCmd(event *irc.Event, callback func(*core.ReplyCallbackData)) bool {
	fields := strings.Fields(event.Message())
	// fields[0]  => Command
	// fields[1] => url for the picture
	// fields[2:] => Tag for the picture
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

	url := fields[1]
	tag := strings.ToLower(strings.Join(fields[2:], " "))

	result, err := dbPtr.Exec(sqlDelete, tag, url)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlDelete)
	}
	rowCount, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlDelete)
	}
	if rowCount != 0 {
		callback(&core.ReplyCallbackData{
			Message: fmt.Sprintf("Picture \"%s\" removed for tag \"%s\"", url, tag),
			Target:  core.GetTargetFromEvent(event)})
	}
	return true
}

func prepareTagString(str string) string {
	return strings.TrimSpace(strings.ToLower(reSanitize.ReplaceAllString(str, "")))
}
