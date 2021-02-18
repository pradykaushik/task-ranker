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
	// Provide this name as the strategy when configuring the task ranker to use it for
	// ranking tasks.
	cpuSharesStrategy = "cpushares"
	// cpuUtilStrategy is the name of the task ranking strategy that ranks
	// tasks in non-increasing order based on the cpu utilization in the past N seconds.
	// Provide this name as the strategy when configuring the task ranker to use it for
	// ranking tasks.
	cpuUtilStrategy         = "cpuutil"
	dynamicToleranceProfile = "dT-profile"
)

var availableStrategies = map[string]strategies.Interface{
	cpuSharesStrategy:       new(strategies.TaskRankCpuSharesStrategy),
	cpuUtilStrategy:         new(strategies.TaskRankCpuUtilStrategy),
	dynamicToleranceProfile: new(strategies.DynamicToleranceProfiler),
}

// GetTaskRankStrategy returns the task ranking strategy with the given name.
func GetTaskRankStrategy(strategy string) (strategies.Interface, error) {
	if s, ok := availableStrategies[strategy]; !ok {
		return nil, errors.New("invalid task ranking strategy")
	} else {
		return s, nil
	}
}
