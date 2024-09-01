package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

type LinkRepository interface {
	RelativePaths(baseUrl string) ([]Link, error)
	UpdateStatus(link Link) (int64, error)
	Add(link Link) (int64, error)
}

type PgxLinkRepository struct { //Pgx Concurrency
	Pool *pgxpool.Pool
}

func NewPgxLinkRepository(pool *pgxpool.Pool) *PgxLinkRepository {
	return &PgxLinkRepository{Pool: pool}
}

func (pgxLinkRepository PgxLinkRepository) RelativePaths(baseUrl string) ([]Link, error) {
	ctx := context.Background()
	conn, err := pgxLinkRepository.Pool.Acquire(ctx)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer conn.Release()

	var links []Link

	query := `
		SELECT id, url, base_url 
		FROM html_link_parser.link 
		WHERE base_url = $1 
		AND url LIKE '/%' 
		AND status_code IS NULL
	`

	// remember to add prepared params!!!
	rows, err := conn.Query(ctx, query, baseUrl)
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

func (pgxLinkRepository PgxLinkRepository) UpdateStatus(link Link) (int64, error) {
	ctx := context.Background()
	conn, err := pgxLinkRepository.Pool.Acquire(ctx)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer conn.Release()

	result, err := conn.Exec(ctx, "UPDATE html_link_parser.link SET status_code = $1, status_message = $2 WHERE id = $3",
		link.StatusCode, link.StatusMessage, link.ID,
	)

	if err != nil {
		return 0, fmt.Errorf("UpdateStatus: %v", err)
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, nil
}

func (pgxLinkRepository PgxLinkRepository) Add(link Link) (int64, error) {
	ctx := context.Background()
	conn, err := pgxLinkRepository.Pool.Acquire(ctx)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer conn.Release()

	query := `
		INSERT INTO html_link_parser.link (url, description, source_url, base_url, created_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`

	var id int64
	queryErr := conn.QueryRow(
		ctx,
		query,
		link.Href, link.Text, link.SourceUrl, link.BaseUrl, time.Now(),
	).Scan(&id)

	if queryErr != nil {
		return 0, fmt.Errorf("Add: %v", err)
	}

	return id, nil
}
