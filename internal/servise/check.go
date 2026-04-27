package servise

import (
	"fmt"
	"time"
	"watchdog/internal/storage"
)

type Service struct {
	storage *storage.Storage
}

type Statistic struct {
	Url                   string `json:"url"`
	SuccessfullyCount     int    `json:"succesfull_count"`
	ErrorCount            int    `json:"error_count"`
	AverageResponsetTimme string `json:"asverageResponsetTimme"`
}

func New(storage *storage.Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func newStatistic(url string) *Statistic {
	return &Statistic{
		Url:                   url,
		SuccessfullyCount:     0,
		ErrorCount:            0,
		AverageResponsetTimme: "",
	}
}

func (s *Service) GetCheckStatistic(url string) (*Statistic, error) {
	chanAverange := make(chan int64)
	chanAverangeRes := make(chan string)
	chanError := make(chan error, 1)
	data, err := s.storage.GetAllDataByUrl(url)
	if err != nil {
		return nil, err
	}

	statistic := newStatistic(url)

	go gettingTheAverange(chanAverange, chanAverangeRes, len(data))

	go func() {
		defer close(chanAverange)
		for _, d := range data {
			diffTime, err := parceAndDifferenceData(d.RequestAt, d.ResponseAt)
			if err != nil {
				chanError <- err
				return
			}
			chanAverange <- diffTime
		}
	}()

	for _, d := range data {
		if d.Status == 200 {
			statistic.SuccessfullyCount++
		}
		if d.ErrorText != "" {
			statistic.ErrorCount++
		}
	}

	select {
	case err := <-chanError:
		if err != nil {
			return nil, err
		}
	case res := <-chanAverangeRes:
		statistic.AverageResponsetTimme = res
	}

	return statistic, nil
}

func gettingTheAverange(data <-chan int64, res chan string, count int) {
	defer close(res)
	var sumData int64
	for d := range data {
		sumData += d
	}
	res <- fmt.Sprintf("%.3f", float64(sumData)/float64(count))
}

func parceAndDifferenceData(req, resp string) (int64, error) {
	layout := "2006-01-02 15:04:05"
	t1, err := time.Parse(layout, req)
	if err != nil {
		return -1, err
	}
	t2, err := time.Parse(layout, resp)
	if err != nil {
		return -1, err
	}
	difference := t2.Unix() - t1.Unix()

	return difference, nil

}
