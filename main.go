package main

import (
	"fmt"
	"my-orchestrator-with-go/task"

	"github.com/google/uuid"
)

func main() {
	t := task.Task{
		ID:     uuid.New(),
		Name:   "Task-1",
		State:  task.Pending,
		Image:  "Image-1",
		Memory: 1024,
		Disk:   1,
	}

	fmt.Printf("task: %v\n", t)
}
