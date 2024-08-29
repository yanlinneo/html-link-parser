package repository

import (
	"database/sql"
	"fmt"
	pkglink "html-link-parser/link"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB() error {
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "html_link_parser",
		ParseTime:            true, //instructs our driver to convert SQL TIME and DATE fields to Go time.Time objects.
		AllowNativePasswords: true,
	}

	slog.Info("Connecting to database...")

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
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

func RelativePaths(baseUrl string) ([]pkglink.Link, error) {
	var links []pkglink.Link

	// remember to add prepared params!!!
	rows, err := db.Query(
		`SELECT id, url, base_url FROM link WHERE base_url LIKE ? AND url LIKE '/%' AND status_code IS NULL`,
		baseUrl,
	)
	if err != nil {
		return nil, fmt.Errorf("RelativePaths: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var link pkglink.Link
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

func UpdateStatus(link pkglink.Link) (int64, error) {
	fmt.Println(link.StatusCode, link.StatusMessage)
	result, err := db.Exec("UPDATE link SET status_code = ?, status_message = ? WHERE id = ?", link.StatusCode, link.StatusMessage, link.ID)
	if err != nil {
		return 0, fmt.Errorf("UpdateStatus: %v", err)
	}

	id, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("UpdateStatus: %v", err)
	}

	return id, nil
}

func Add(link pkglink.Link) (int64, error) {
	result, err := db.Exec("INSERT INTO link (url, description, source_url, base_url, created_at) VALUES (?, ?, ?, ?, ?)", link.Href, link.Text, link.SourceUrl, link.BaseUrl, time.Now())
	if err != nil {
		return 0, fmt.Errorf("Add: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Add: %v", err)
	}

	return id, nil
}
