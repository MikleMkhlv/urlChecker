package main

import (
	"fmt"
	"log"
	"net/http"
	"watchdog/internal/cheacker"
	"watchdog/internal/handlers"
	"watchdog/internal/queue"
	"watchdog/internal/scheduler"
	"watchdog/internal/servise"
	"watchdog/internal/storage"
)

func main() {
	connStr := "postgres://itb2:itb2@10.32.200.168:5432/watchdog?sslmode=disable"

	s, err := storage.NewStorage(connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД %v\n", err)
	}
	fmt.Println("Успешное подключение")

	//RabbitMQ
	brConn, err := queue.New("amqp://guest:guest@10.32.200.168:5672/", s)
	if err != nil {
		log.Fatalf("Не удалось подключиться к RabbitMQ %v\n", err)
	}
	producer, err := brConn.NewProducer("sites_queue")
	if err != nil {
		log.Fatalf("Ошибка создания продюсера %v\n", err)
	}

	// ==========Worker============
	// jobs := make(chan storage.Sites, 100)
	worker := cheacker.New(s)

	for i := 1; i <= 3; i++ {
		consumer, err := brConn.NewConsumer("sites_queue")
		if err != nil {
			log.Fatal(err)
		}
		go consumer.Consume(i, worker.HandleSite)
	}

	stop := make(chan struct{})
	go scheduler.RunShedualer(s, producer, stop)

	serviceStat := servise.New(s)

	handlers := handlers.NewHandlers(s, serviceStat)

	http.HandleFunc("/sites", handlers.Sites)
	http.HandleFunc("/checks", handlers.CheckStatiscs)

	fmt.Println("Started server with зщке 8080.")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
