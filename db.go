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

func (db *Database) CreateTable(table string) (status bool) {
	stmt := `CREATE TABLE $1 (
				id SERIAL PRIVATE KEY,
				resource_id TEXT NOT NULL,
				ttl TIMESTAMP NOT NULL
			);`
	_, err := db.database.Exec(stmt, table)
	if err == nil {
		status = true
	}
	return
}

func (db *Database) DropTable(table string) (status bool) {
	stmt := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
	_, err := db.database.Exec(stmt)
	if err == nil {
		status = true
	}
	return
}

func (db *Database) TableExists(table string) (exists bool) {
	tbl, _ := db.database.Query(`SELECT COUNT(table_name)
		FROM
			information_schema.tables
		WHERE
			table_schema LIKE 'public' AND
			table_type LIKE 'BASE TABLE' AND
			table_name = $1;`, table)
	defer tbl.Close()
	count := 0
	for tbl.Next() {
		tbl.Scan(&count)
		break
	}
	if count == 1 {
		exists = true
	}
	return
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
