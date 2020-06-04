package entities

type Task struct {
	// CpuShares allocated to the task.
	CpuShares float64
	// Memory allocated to the task.
	RAM float64
	// Labels assigned to the docker container inside which the task is running.
	Labels map[string]string
}
