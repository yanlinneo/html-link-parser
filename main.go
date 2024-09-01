package main

import (
	"context"
	"flag"
	"fmt"
	"html-link-parser/models"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var baseUrl string
var pool *pgxpool.Pool

func InitDB() error {
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	dbName := "go_app"
	var err error

	slog.Info("Connecting to database...")
	connStr := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", dbUser, dbPass, dbName)
	pool, err = pgxpool.New(context.Background(), connStr)

	if err != nil {
		log.Fatal(err)
		return err
	}

	slog.Info("Database is connected!")
	return nil
}

func main() {
	var urlFlag = flag.String("url", "", "Enter a URL")
	flag.Parse()

	if validateErr := Validate(urlFlag); validateErr != nil {
		fmt.Println(validateErr)
		return
	}

	fmt.Println("Processing", *urlFlag)

	start := time.Now()
	var err error

	// start the program with 1 link (sourceLink)
	var sourceLink = models.Link{Href: *urlFlag}

	// add sourceLink to pendingLinks
	var pendingLinks []models.Link
	pendingLinks = append(pendingLinks, sourceLink)

	// start db
	dbErr := InitDB()
	if dbErr != nil {
		slog.Error("Failed to initialize DB:", "error", dbErr)
		os.Exit(1)
	}

	linkRepo := models.NewPgxLinkRepository(pool)

	// Ensure the pool is closed when the program exits
	defer pool.Close()

	// pendingLinks will be processed in this loop
	for {
		concurrentProcess(pendingLinks, linkRepo)

		// fetch all relative paths (pendingLinks) from the database
		pendingLinks, err = linkRepo.RelativePaths(baseUrl)
		if err != nil {
			slog.Error("Relative Path Links:", "error", err)
		}

		// if there are no more pendingLinks to process, break
		if len(pendingLinks) == 0 {
			break
		}
	}

	fmt.Println("Total Duration: ", time.Since(start))
}
