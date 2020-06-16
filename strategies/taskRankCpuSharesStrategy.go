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
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/entities"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/prometheus/common/model"
	"sort"
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
func (s *TaskRankCpuSharesStrategy) Execute(data model.Value) {
	valueT := data.Type()
	var matrix model.Matrix
	// Safety check to make sure that we cast to matrix only if value type is matrix.
	// Note, however, that as the strategy decides the metric and the range for fetching
	// data, it can assume the value type.
	// For example, if a range is provided, then the value type would
	// be a matrix.
	switch valueT {
	case model.ValMatrix:
		matrix = data.(model.Matrix)
	default:
		// invalid value type.
		// TODO do not ignore this. maybe log it?
	}

	// Initializing tasks to rank.
	var tasks []entities.RankedTask
	for _, sampleStream := range matrix {
		tasks = append(tasks, entities.RankedTask{
			Metric: sampleStream.Metric,
			// As cpu shares allocated to a container can be updated for docker containers,
			// taking the average of allocated cpu shares.
			Weight: s.avgCpuShare(sampleStream.Values),
		})
	}

	// Sorting the tasks in non-increasing order of cpu shares.
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].Weight > tasks[j].Weight
	})

	// Submitting the ranked tasks to the receiver.
	s.receiver.Receive(tasks)
}

// avgCpuShare returns the average cpushare allocated to a container.
func (s TaskRankCpuSharesStrategy) avgCpuShare(values []model.SamplePair) float64 {
	sum := 0.0
	for _, val := range values {
		sum += float64(val.Value)
	}
	return sum / float64(len(values))
}

// GetMetric returns the name of the metric to query.
func (s TaskRankCpuSharesStrategy) GetMetric() string {
	// TODO convert this to constant.
	return "container_spec_cpu_shares"
}

// SetLabelMatchers sets the label matchers to use to filter data.
func (s *TaskRankCpuSharesStrategy) SetLabelMatchers(labelMatchers []*query.LabelMatcher) error {
	if len(labelMatchers) == 0 {
		return errors.New("no label matchers provided")
	}
	s.labels = labelMatchers
	return nil
}

// GetLabelMatchers returns the label matchers to be used to filter data.
func (s TaskRankCpuSharesStrategy) GetLabelMatchers() []*query.LabelMatcher {
	return s.labels
}

// GetRange returns the time unit and duration for how far back values need to be fetched.
func (s TaskRankCpuSharesStrategy) GetRange() (query.TimeUnit, uint) {
	return query.Seconds, 5
}
