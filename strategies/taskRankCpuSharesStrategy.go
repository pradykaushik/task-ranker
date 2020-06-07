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

package strategies

import (
	"github.com/pradykaushik/task-ranker/query"
	"log"
)

// TaskRankCpuSharesStrategy is a task ranking strategy that ranks the tasks
// in non-increasing order based on the cpu-shares allocated to tasks.
type TaskRankCpuSharesStrategy struct {
	// receiver of the results of task ranking.
	receiver TaskRanksReceiver
	// labels used to filter the time series data fetched from prometheus.
	labels []*query.LabelMatcher
}

// SetTaskRanksReceiver sets the receiver of the results of task ranking.
func (s *TaskRankCpuSharesStrategy) SetTaskRanksReceiver(receiver TaskRanksReceiver) {
	s.receiver = receiver
}

// Execute the strategy using the provided data.
func (s *TaskRankCpuSharesStrategy) Execute(data string) {
	// placeholder.
	s.receiver.Receive(nil)
	log.Println("from::cpushares_strategy with query string::" + data)
}

// GetMetric returns the name of the metric to query.
func (s TaskRankCpuSharesStrategy) GetMetric() string {
	// TODO convert this to constant.
	return "container_spec_cpu_shares"
}

// SetLabelMatchers sets the label matchers to use to filter data.
func (s *TaskRankCpuSharesStrategy) SetLabelMatchers(labelMatchers []*query.LabelMatcher) {
	s.labels = labelMatchers
}

// GetLabelMatchers returns the label matchers to be used to filter data.
func (s TaskRankCpuSharesStrategy) GetLabelMatchers() []*query.LabelMatcher {
	return s.labels
}

// GetRange returns the time unit and duration for how far back values need to be fetched.
func (s TaskRankCpuSharesStrategy) GetRange() (query.TimeUnit, int) {
	return query.Seconds, 5
}
