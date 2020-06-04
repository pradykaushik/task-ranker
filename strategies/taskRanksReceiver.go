package strategies

import "github.com/pradykaushik/task-ranker/entities"

type TaskRanksReceiver interface {
	// Receive the ranked tasks and their corresponding information.
	// Task at position i is ranked higher than task at position i+1.
	Receive([]entities.Task)
}
