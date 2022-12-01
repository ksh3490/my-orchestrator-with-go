package main

import (
	"log"
	"my-orchestrator-with-go/task"
	"my-orchestrator-with-go/worker"
	"os"
	"strconv"
	"time"

	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	host := os.Getenv("MY_HOST")
	port, _ := strconv.Atoi(os.Getenv("MY_PORT"))

	fmt.Printf("Starting Worker in %s:%d\n", host, port)

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	api := worker.Api{Address: host, Port: port, Worker: &w}

	go runTasks(&w)
	go w.CollectStats()
	api.Start()
}

func runTasks(w *worker.Worker) {
	for {
		if w.Queue.Len() != 0 {
			result := w.RunTask()
			if result.Error != nil {
				log.Printf("Error running task: %v\n", result.Error)
			}
		} else {
			log.Printf("No Task to process currently.\n")
		}
		log.Println("Sleeping for 10 seconds.")
		time.Sleep(10 * time.Second)
	}
}
