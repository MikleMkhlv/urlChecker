package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Sites struct {
	Id            int
	Url           string
	LastCheckedAt time.Time
	Interval      int
	IsDQL         bool
}

func NewSites(id int, url string, lastime time.Time, interval int, isDLQ bool) (*Sites, error) {
	if url == "" {
		return nil, fmt.Errorf("url at id: %d is empty", id)
	}
	return &Sites{
		Id:            id,
		Url:           url,
		LastCheckedAt: lastime,
		Interval:      interval,
		IsDQL:         isDLQ,
	}, nil
}

func (s *Storage) SaveSite(url string) error {
	timeNow := time.Now()
	_, err := s.db.Exec("INSERT INTO sites (url, last_checked_at) VALUES ($1, $2)", url, timeNow)
	var customErr error
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			switch pgxErr.Code {
			case "23505":
				customErr = fmt.Errorf("url {%s} already exists\nerr: %v\n", url, err)
			case "23502":
				customErr = fmt.Errorf("a required field is missing\nerr: %v\n", err)
			default:
				customErr = err
			}
		}
		return customErr
	}
	return nil

}

func (s *Storage) SaveSites(urls []string) []error {
	var sqlErrors []error
	for _, url := range urls {
		err := s.SaveSite(url)
		sqlErrors = append(sqlErrors, err)
	}

	return sqlErrors
}

func (s *Storage) GetAllSites() ([]Sites, []error) {
	var errs []error
	rows, err := s.db.Query("SELECT * FROM sites WHERE last_checked_at <= $1 AND is_DLQ = false ORDER BY last_checked_at LIMIT 100", time.Now())

	if err != nil {
		return nil, append(make([]error, 1), err)
	}
	defer rows.Close()

	var sites []Sites
	for rows.Next() {
		var id int
		var url string
		var lastCheckedAt time.Time
		var interval int
		var isDLQ bool
		if err := rows.Scan(&id, &url, &lastCheckedAt, &interval, &isDLQ); err != nil {
			errs = append(errs, err)
			continue
		}
		site, err := NewSites(id, url, lastCheckedAt, interval, isDLQ)
		if err != nil {
			errs = append(errs, err)
		}
		sites = append(sites, *site)
	}
	return sites, errs
}

func (s *Storage) UpdateSites(st Sites) error {
	updatedLastCheckedTime := time.Now().Add(time.Duration(st.Interval))
	_, err := s.db.Exec("UPDATE sites SET last_checked_at = $1 WHERE id = $2", updatedLastCheckedTime, st.Id)
	return err
}

func (s *Storage) UpdateDLQStatus(st Sites, isDLQ bool) error {
	_, err := s.db.Exec("UPDATE sites SET is_DLQ = $1 WHERE id = $2", isDLQ, st.Id)
	return err
}
