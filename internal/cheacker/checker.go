package cheacker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"watchdog/internal/storage"
)

type Worker struct {
	store *storage.Storage
}

func New(stopre *storage.Storage) *Worker {
	return &Worker{
		store: stopre,
	}
}

// func Worker(id int, jobs <-chan storage.Sites, storage *storage.Storage) {
// 	httpClient := http.Client{
// 		Timeout: time.Second * 5,
// 	}

// 	for site := range jobs {
// 		fmt.Printf("[Worker-%d] is cheack url: %v\n", id, site)
// 		// beginCurrentTime := time.Now().Format("2006-01-02 15:04:05")
// 		beginCurrentTime := time.Now()
// 		resp, err := httpClient.Get(site.Url)
// 		if err != nil {
// 			fmt.Printf("[Worker-%d]: Error check url: %v\n", id, site)
// 			err := storage.SaveCheck(site.Url, 0, err.Error(), beginCurrentTime)
// 			if err != nil {
// 				fmt.Println(err.Error())
// 			}
// 			continue
// 		}
// 		resp.Body.Close()

// 		err = storage.SaveCheck(site.Url, resp.StatusCode, "", beginCurrentTime)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		}
// 		err = storage.UpdateSites(site)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		}

// 		fmt.Printf("[Worker-%d] check url: %s success\n", id, site.Url)
// 	}
// }

func (w *Worker) HandleSite(workerId int, body []byte) error {
	var site storage.Sites
	fmt.Println("HANDLER START")

	err := json.Unmarshal(body, &site)
	if err != nil {
		return err
	}
	fmt.Printf("[Worker-%d] is cheack url: %s\n", workerId, site.Url)

	httpClient := http.Client{
		Timeout: time.Second * 5,
	}

	beginCurrentTime := time.Now()
	resp, err := httpClient.Get(site.Url)
	if err != nil {
		fmt.Printf("[Worker-%d]: Error check url: %s\n", workerId, site.Url)
		w.store.SaveCheck(site.Url, 0, err.Error(), beginCurrentTime)
		return err
	}
	defer resp.Body.Close()

	err = w.store.SaveCheck(site.Url, resp.StatusCode, "", beginCurrentTime)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = w.store.UpdateSites(site)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("[Worker-%d] check url: %s success\n", workerId, site.Url)
	fmt.Println("HANDLER END")

	return nil
}
