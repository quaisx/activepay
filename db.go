package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost" //docker PostgreSql is running on -p 5432:5432
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "active"
)

type Database struct {
	database *sql.DB
}

func (db *Database) Update(m map[string]time.Time) {
	tx, _ := db.database.Begin()
	stmt, _ := tx.Prepare("UPDATE active SET ttl = $1 WHERE resource_id = $2; ")
	defer stmt.Close()
	for res, ttl := range m {
		if !ttl.IsZero() {
			stmt.Exec(res, ttl)
		}
	}
	tx.Commit()
	stmt.Close()
}

func (db *Database) Select(resource_id string) time.Time {
	// query
	rows, _ := db.database.Query("SELECT ttl FROM active WHERE resource_id = $1", resource_id)
	defer rows.Close()

	ttl := time.Time{}
	for rows.Next() {
		rows.Scan(&ttl)
		break
	}
	return ttl
}

func (db *Database) NewConnection() {
	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db.database, _ = sql.Open("postgres", psqlconn)
}

func (db *Database) Close() {
	db.Close()
}
