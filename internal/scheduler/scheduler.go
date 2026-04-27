package scheduler

import (
	"encoding/json"
	"fmt"
	"time"
	"watchdog/internal/queue"
	"watchdog/internal/storage"
)

func RunShedualer(store *storage.Storage, producer *queue.Producer, stop <-chan struct{}) {
	ticket := time.NewTicker(time.Second * 10)
	defer ticket.Stop()

	for {
		select {
		case <-ticket.C:
			sites, err := store.GetAllSites()
			if err != nil {
				continue
			}

			for _, site := range sites {

				body, err := json.Marshal(site)
				if err != nil {
					fmt.Println("error marshal sittes")
				}
				err = producer.PublishMessage(body)

			}
		case <-stop:
			fmt.Println("Stop scheduler")
			return
		}
	}
}
