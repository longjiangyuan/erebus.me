package model

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func SetDataSource(source string) {
	_db, err := sql.Open("mysql", source)
	if err != nil {
		log.Fatal(err)
	}
	db = _db
}

func beginTransaction(fn func(*sql.Tx)) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
	fn(tx)
}

func Query(statement string, args ...interface{}) *sql.Rows {
	rows, err := db.Query(statement, args...)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	return rows
}

func ExecUpdate(statement string, args ...interface{}) int64 {
	rs, err := db.Exec(statement, args...)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	affected, err := rs.RowsAffected()
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	return affected
}
