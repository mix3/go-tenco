package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type storage struct {
	db *sql.DB
}

func New(dsn string) (*storage, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	sql := `
		CREATE TABLE IF NOT EXISTS subdomain_map (
			subdomain TEXT PRIMARY KEY,
			backend   TEXT
		)
	`
	_, err = db.Exec(sql)
	if err != nil {
		return nil, err
	}
	return &storage{db: db}, nil
}

func (s *storage) Map() (map[string]string, error) {
	rows, err := s.db.Query("SELECT subdomain, backend FROM subdomain_map")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := map[string]string{}
	for rows.Next() {
		var key, val string
		rows.Scan(&key, &val)
		ret[key] = val
	}
	return ret, nil
}

func (s *storage) Get(key string) (string, error) {
	var val string
	err := s.db.
		QueryRow("SELECT backend FROM subdomain_map WHERE subdomain = ?", key).
		Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *storage) Set(key, val string) error {
	sql := `
		REPLACE INTO subdomain_map (subdomain, backend)
		VALUES (?, ?)
	`
	_, err := s.db.Exec(sql, key, val)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM subdomain_map WHERE subdomain = ?", key)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) DeleteAll() error {
	_, err := s.db.Exec("DELETE FROM subdomain_map")
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) Close() {
	s.db.Close()
}
