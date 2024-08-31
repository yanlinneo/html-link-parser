package main

import (
	"flag"
	"fmt"
	"html-link-parser/models"
	"log/slog"
	"time"
)

var baseUrl string

func main() {
	var urlFlag = flag.String("url", "", "help")
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
	dbErr := models.InitDB()
	if dbErr != nil {
		slog.Error("Failed to initialize DB:", "error", dbErr)
		return
	}

	// pendingLinks will be processed in this loop
	for {
		concurrentProcess(pendingLinks)

		// fetch all relative paths (pendingLinks) from the database
		pendingLinks, err = models.RelativePaths(baseUrl)
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
