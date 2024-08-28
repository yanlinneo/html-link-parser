package repository

import (
	"database/sql"
	"fmt"
	pkglink "html-link-parser/link"
	"log"
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
		ParseTime:            true,
		AllowNativePasswords: true,
	}

	//parseTime=true part of the DSN above is a driver-specific parameter which instructs our driver to convert SQL TIME and DATE fields to Go time.Time objects.
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

	fmt.Println("Connected!")
	return nil
}

func AllLinks() ([]pkglink.Link, error) {
	var links []pkglink.Link

	rows, err := db.Query("SELECT * FROM links")
	if err != nil {
		return nil, fmt.Errorf("Links: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var link pkglink.Link
		if err := rows.Scan(&link.ID, &link.Href, &link.Text, &link.CreatedDateTime); err != nil {
			return nil, fmt.Errorf("Links: %v", err)
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Links: %v", err)
	}

	return links, nil
}

func AddLink(link pkglink.Link) (int64, error) {
	result, err := db.Exec("INSERT INTO link (url, description, source_url, created_at) VALUES (?, ?, ?, ?)", link.Href, link.Text, link.SourceUrl, time.Now())
	if err != nil {
		return 0, fmt.Errorf("AddLink: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("AddLink: %v", err)
	}

	return id, nil
}
