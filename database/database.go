// The MIT License (MIT)
//
// Copyright (c) 2015 Arnaud Vazard
//
// See LICENSE file.

// Package database manages the database used by the bot
package database

import (
	"database/sql"
	"errors"
	// Imported blank because we only use the driver
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var dbPtr *sql.DB

// NewDatabase creates a new database.
// If database name is an empty string the default path will be used ("./storage/db.sqlite"),
//else it will be used as the path for the database file.
// If reset is true destroy the database before opening it (which will recreate it).
func NewDatabase(databaseName string, reset bool) *sql.DB {
	// Use default name if not specified
	if databaseName == "" {
		// check if the storage directory exist, if not create it
		storage, err := os.Stat("./storage")
		if err != nil {
			os.Mkdir("./storage", os.ModeDir)
		} else if !storage.IsDir() {
			// check if the storage is indeed a directory or not
			log.Fatal("\"storage\" exist but is not a directory")
		}
		databaseName = "./storage/db.sqlite"
	}
	if reset {
		os.Remove(databaseName)
	}
	db, err := sql.Open("sqlite3", databaseName)
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `CREATE TABLE IF NOT EXISTS User (
    nick TEXT NOT NULL PRIMARY KEY,
    email TEXT);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	dbPtr = db
	return db
}

// AddUser adds an user to the database.
func AddUser(nick, email string) (err error) {
	if dbPtr == nil {
		return errors.New("Database pointer is nil")
	}
	sqlStmt := `INSERT OR REPLACE INTO User VALUES ($1, $2)`
	_, err = dbPtr.Exec(sqlStmt, nick, email)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}
