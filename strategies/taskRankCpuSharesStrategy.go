package strategies

import (
	"log"
)

// TaskRankCpuSharesStrategy is a task ranking strategy that ranks the tasks
// in non-increasing order based on the cpu-shares allocated to tasks.
type TaskRankCpuSharesStrategy struct {
	receiver TaskRanksReceiver
}

// SetTaskRanksReceiver sets the receiver of the results of task ranking.
func (s *TaskRankCpuSharesStrategy) SetTaskRanksReceiver(receiver TaskRanksReceiver) {
	s.receiver = receiver
}

func (s *TaskRankCpuSharesStrategy) Execute() {
	// placeholder.
	s.receiver.Receive(nil)
	log.Println("cpu-share task ranking running...")
}
