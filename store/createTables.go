package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func createTables() {
	db, err := sql.Open("sqlite3", "db/bookstore.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS "books" (
		"title"	TEXT,
		"description"	TEXT,
		"isbn"	TEXT,
		"source"	INTEGER NOT NULL,
		"id"	INTEGER NOT NULL UNIQUE,
		PRIMARY KEY("id" AUTOINCREMENT)
		CONSTRAINT "title_from_source" UNIQUE("title","source")
	);
	CREATE TABLE IF NOT EXISTS "images" (
		"name"	TEXT NOT NULL UNIQUE,
		"source"	INTEGER,
		"image"	BLOB NOT NULL,
		"bookId"	INTEGER,
		"hash"	TEXT NOT NULL,
		UNIQUE("source","hash"),
		FOREIGN KEY("bookId") REFERENCES "books"("id")
	);
	`)
	if err != nil {
		panic(err)
	}
}

func init() {
	createTables()
}
