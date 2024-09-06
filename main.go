package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"html-link-parser/models"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var baseUrl string
var baseUrlHost string
var relativePathsAccessed int
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
	var urlFlag = flag.String("url", "", "Enter an URL:")
	flag.Parse()

	if validateErr := Validate(urlFlag); validateErr != nil {
		fmt.Println(validateErr)
		return
	}

	slog.Info("Processing", "URL", *urlFlag)

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

	//print csv
	dbLinks, err := linkRepo.AllLinksFrom(baseUrl)
	if err != nil {
		fmt.Println(err)
	}

	totalLinks := len(dbLinks)
	linkRecords := [][]string{}

	// headers
	headers := []string{"Relative Paths / URL", "Text", "Status Code", "Status Message"}
	linkRecords = append(linkRecords, headers)

	// rows/records
	for _, dbL := range dbLinks {
		record := []string{dbL.Href, dbL.Text, strconv.Itoa(int(dbL.StatusCode.Int32)), dbL.StatusMessage.String}
		linkRecords = append(linkRecords, record)
	}

	currentTime := time.Now()
	timeString := fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	fileName := baseUrlHost + "_" + timeString + ".csv"

	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("failed to create file:", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.WriteAll(linkRecords)

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

	slog.Info("Created", "File", fileName)
	slog.Info("Results",
		"Total Links Saved", totalLinks,
		"Relative Paths Accessed", relativePathsAccessed,
		"Duration", time.Since(start))
}
