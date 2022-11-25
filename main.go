package main

import (
	"log"
	"my-orchestrator-with-go/task"
	"my-orchestrator-with-go/worker"
	"net"
	"os"
	"strings"
	"time"

	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	hostname, _ := os.Hostname()
	host, _ := net.LookupHost(hostname)
	hostIp := strings.Join(host, "")
	port := 5555

	fmt.Println("Starting Worker")

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	api := worker.Api{Address: hostIp, Port: port, Worker: &w}

	go runTasks(&w)
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
