package service

import (
	"database/sql"
)

var db *sql.DB

func IntilizeDb(Db *sql.DB) {
	db = Db
}
