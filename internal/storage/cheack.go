package storage

import "time"

const DATA_FORMAT string = "2006-01-02 15:04:05"

type Check struct {
	Id         int
	Url        string
	Status     int
	ErrorText  string
	RequestAt  string
	ResponseAt string
}

func NewCheck(id int, url string, status int, errorText string, requestAt time.Time, responseAt time.Time) *Check {
	return &Check{
		Id:         id,
		Url:        url,
		Status:     status,
		ErrorText:  errorText,
		RequestAt:  requestAt.Format(DATA_FORMAT),
		ResponseAt: responseAt.Format(DATA_FORMAT),
	}
}

func (s *Storage) SaveCheck(url string, status int, errorText string, beginRequestTime time.Time) error {
	afterCurrentTime := time.Now()
	_, err := s.db.Exec("INSERT INTO checkStatus (url, status, error, requestAt, responseAt) VALUES ($1, $2, $3, $4, $5)", url, status, errorText, beginRequestTime, afterCurrentTime)
	return err
}

func (s *Storage) GetAllDataByUrl(url string) ([]Check, error) {
	rows, err := s.db.Query("SELECT * FROM checkStatus")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allData []Check
	for rows.Next() {
		var id, status int
		var url, errorT string
		var requestAt, responseAt time.Time
		if err := rows.Scan(&id, &url, &status, &errorT, &requestAt, &responseAt); err != nil {
			continue
		}
		check := NewCheck(id, url, status, errorT, requestAt, responseAt)
		allData = append(allData, *check)
	}
	return allData, nil
}
