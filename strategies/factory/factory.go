package factory

import (
	"errors"
	"github.com/pradykaushik/task-ranker/strategies"
)

const (
	// cpuSharesStrategy is the name of the task ranking strategy that ranks
	// tasks in non-increasing order based on the allocated cpu-shares.
	// Provide this name as the strategy in the task ranker config to use this strategy
	// for ranking tasks.
	cpuSharesStrategy = "cpushares"
)

var availableStrategies = map[string]struct{}{
	cpuSharesStrategy: {},
}

// GetTaskRankStrategy instantiates a new task ranking strategy.
func GetTaskRankStrategy(strategy string) (strategies.Strategy, error) {
	if _, ok := availableStrategies[strategy]; !ok {
		return nil, errors.New("invalid task ranking strategy")
	}
	return new(strategies.TaskRankCpuSharesStrategy), nil
}
