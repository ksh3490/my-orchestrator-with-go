package task

type State int

const (
	Pending State = iota
	Scheduled
	Completed
	Running
	Failed
)
