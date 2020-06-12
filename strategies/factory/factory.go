// Copyright 2020 Pradyumna Kaushik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package factory

import (
	"github.com/pkg/errors"
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
func GetTaskRankStrategy(strategy string) (strategies.Interface, error) {
	if _, ok := availableStrategies[strategy]; !ok {
		return nil, errors.New("invalid task ranking strategy")
	}
	return new(strategies.TaskRankCpuSharesStrategy), nil
}
