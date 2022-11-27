package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"my-orchestrator-with-go/task"

	linux "github.com/c9s/goprocinfo/linux"
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

type Stats struct {
	MemStats  *linux.MemInfo
	DiskStats *linux.Disk
	CpuStats  *linux.CPUStat
	LoadStats *linux.LoadAvg
}

func (w *Worker) CollectStats() {
	fmt.Println("I will collect stats")
}

func (s *Stats) MemTotalKb() uint64 {
	return s.MemStats.MemTotal
}

func (s *Stats) MemAvailableKb() uint64 {
	return s.MemStats.MemAvailable
}

func (s *Stats) MemUsedKb() uint64 {
	return s.MemStats.MemTotal - s.MemStats.MemAvailable
}

func (s *Stats) MemUsedPercent() uint64 {
	return s.MemStats.MemAvailable / s.MemStats.MemTotal
}

func (w *Worker) RunTask() task.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("No task in the queue")
		return task.DockerResult{Error: nil}
	}

	taskQueued := t.(task.Task)

	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = &taskQueued
	}

	var result task.DockerResult
	if task.ValidStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			result = w.StartTask(taskQueued)
			w.TaskCount += 1
		case task.Completed:
			result = w.StopTask(taskQueued)
			w.TaskCount -= 1
		default:
			result.Error = errors.New("we should not get here")
		}
	} else {
		err := fmt.Errorf("invalid transition from %v to %v", taskPersisted.State, taskQueued.State)
		result.Error = err
	}
	return result
}

func (w *Worker) GetTasks() []task.Task {

	result := make([]task.Task, 0, len(w.Db))
	for _, value := range w.Db {
		result = append(result, *value)
	}

	return result
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
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
