package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"urlsh/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New" // Mark for errors

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: to add a timestamp field CreatedAt
	// TODO: to add a number field to count amount of redirects
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT TEXT NOT NULL,
		used_count INTEGER DEFAULT 0
		);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	CREATE TABLE IF NOT EXISTS redirect_analysis(
		id INTEGER PRIMARY KEY,
		used_url INTEGER NOT NULL references url(id),
		used_at TIMESTAMP
	);


	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// TODO: refactor
		if sqlireErr, ok := err.(sqlite3.Error); ok && sqlireErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

//func (s *Storage) GetURL(alias string) (string, error) {
//	const op = "storage.sqlite.GetURL"
//
//	stmt, err := s.db.Prepare(`
//		SELECT url, id FROM url WHERE alias = ?;
//		`)
//	if err != nil {
//		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
//	}
//
//	var resURL string
//	var urlID int64
//	err = stmt.QueryRow(alias).Scan(&resURL, &urlID)
//	if errors.Is(err, sql.ErrNoRows) {
//		return "", storage.ErrUrlNotFound
//	}
//	if err != nil {
//		return "", fmt.Errorf("%s: execute statement: %w", op, err)
//	}
//
//	return resURL, nil
//}

func (s *Storage) GetURLAndLogRedirect(alias string) (string, error) {
	const op = "storage.sqlite.GetURLAndLogRedirect"

	// Get URL and ID
	stmt, err := s.db.Prepare(`
		SELECT url, id FROM url WHERE alias = ?;
		`)
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string
	var urlID int64
	err = stmt.QueryRow(alias).Scan(&resURL, &urlID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrUrlNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	// Log redirect in redirect_analysis
	insertStmt, err := s.db.Prepare("INSERT INTO redirect_analysis(used_url, used_at) VALUES(?, datetime('now'))")
	if err != nil {
		return "", fmt.Errorf("%s: prepare insert statement: %w", op, err)
	}

	_, err = insertStmt.Exec(urlID)
	if err != nil {
		return "", fmt.Errorf("%s: insert redirect log: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrUrlNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: delete error: %w", op, err)
	}

	return nil
}

// TODO: create GetAnalysis function (get used_amount and used time)

func (s *Storage) GetUrlAnalysis(alias string) (int64, []time.Time, error) {
	const op = "storage.sqlite.GetUrlAnalysis"

	// get id parameter of gotten alias
	stmt, err := s.db.Prepare(`
		SELECT used_count, id FROM url WHERE alias = ?;
		`)
	if err != nil {
		return 0, nil, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var urlID int64
	var usedCount int64
	err = stmt.QueryRow(alias).Scan(&usedCount, &urlID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil, storage.ErrUrlNotFound
	}
	if err != nil {
		return 0, nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	rows, err := s.db.Query("SELECT used_at FROM redirect_analysis WHERE used_url = ?;", urlID)
	defer rows.Close()

	if err != nil {
		return usedCount, nil, fmt.Errorf("%s: get used times error: %w", op, err)
	}

	usedTimes := make([]time.Time, 0)
	for rows.Next() {
		var usedTime time.Time
		err = rows.Scan(&usedTime)
		if err != nil {
			return usedCount, nil, fmt.Errorf("%s: handle used times error: %w", op, err)
		}
		usedTimes = append(usedTimes, usedTime)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}

	return usedCount, usedTimes, nil
}
