package models

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Link struct {
	ID              int64
	Href            string
	Text            string
	SourceUrl       string
	BaseUrl         string
	CreatedDateTime time.Time
	StatusCode      int
	StatusMessage   string
}

var db *sql.DB

func InitDB() error {
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	dbName := "go_app"
	var err error

	slog.Info("Connecting to database...")
	connStr := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", dbUser, dbPass, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		db.Close()
		log.Fatal(pingErr)
		return pingErr
	}

	slog.Info("Database is connected!")
	return nil
}

func RelativePaths(baseUrl string) ([]Link, error) {
	var links []Link

	query := `
		SELECT id, url, base_url 
		FROM html_link_parser.link 
		WHERE base_url = $1 
		AND url LIKE '/%' 
		AND status_code IS NULL
	`

	// remember to add prepared params!!!
	rows, err := db.Query(query, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("RelativePaths: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.Href, &link.BaseUrl); err != nil {
			return nil, fmt.Errorf("RelativePaths: %v", err)
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("RelativePaths: %v", err)
	}

	return links, nil
}

func (link Link) UpdateStatus() (int64, error) {
	result, err := db.Exec("UPDATE html_link_parser.link SET status_code = $1, status_message = $2 WHERE id = $3",
		link.StatusCode, link.StatusMessage, link.ID,
	)

	if err != nil {
		return 0, fmt.Errorf("UpdateStatus: %v", err)
	}

	id, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("UpdateStatus: %v", err)
	}

	return id, nil
}

func (link Link) Add() (int64, error) {
	query := `
		INSERT INTO html_link_parser.link (url, description, source_url, base_url, created_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`

	var id int64
	err := db.QueryRow(query, link.Href, link.Text, link.SourceUrl, link.BaseUrl, time.Now()).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("Add: %v", err)
	}

	return id, nil
}
