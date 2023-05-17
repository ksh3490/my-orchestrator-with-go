package main

import (
	"log"
	"my-orchestrator-with-go/manager"
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
	workers := []string{fmt.Sprintf("%s:%d", host, port)}
	m := manager.New(workers)
	api := worker.Api{Address: host, Port: port, Worker: &w}

	for i := 0; i < 3; i++ {
		t := task.Task{
			ID: uuid.New(),
			Name: fmt.Sprintf("test-container-%d", i),
			State: task.Scheduled,
			Image: "strm/helloworld-http",
		}
		te := task.TaskEvent{
			ID: uuid.New(),
			State: task.Running,
			Task: t,
		}
		m.AddTask(te)
		m.SendWork()
	}

	go runTasks(&w)
	go w.CollectStats()
	go api.Start()
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
