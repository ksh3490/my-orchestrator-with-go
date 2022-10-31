package worker

import (
	"fmt"
	"log"
	"time"

	"my-orchestrator-with-go/task"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
}

func (w *Worker) CollectStats() {
	fmt.Println("I will collect stats")
}

func (w *Worker) RunTask() {
	fmt.Println("I will start or stop a task")
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	config := task.Config{
		Name:  t.Name,
		Image: t.Image,
	}

	cli, _ := client.NewClientWithOpts(client.FromEnv)

	d := task.Docker{
		Client:      cli,
		Config:      config,
		ContainerId: t.ContainerID,
	}

	result := d.Run()
	if result.Error != nil {
		log.Printf("Error running task %v: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t

	return result
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.Config{
		Name:  t.Name,
		Image: t.Image,
	}
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	d := task.Docker{
		Client:      cli,
		Config:      config,
		ContainerId: t.ContainerID,
	}
	result := d.Stop()
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v", t.ContainerID, result.Error)
	}
	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container %v for task %v", t.ContainerID, t.ID)

	return result
}
