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
	"github.com/pradykaushik/task-ranker/logger"
	"github.com/pradykaushik/task-ranker/logger/topic"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"sort"
	"time"
)

// TaskRankCpuSharesStrategy is a task ranking strategy that ranks the tasks
// in non-increasing order based on the cpu-shares allocated to tasks.
type TaskRankCpuSharesStrategy struct {
	// receiver of the results of task ranking.
	receiver TaskRanksReceiver
	// labels used to filter the time series data fetched from prometheus.
	labels []*query.LabelMatcher
	// dedicatedLabelNameTaskID is the dedicated label to use when filtering metrics based on task id.
	// Storing this for quick access instead of performing another O(n) search through labels.
	dedicatedLabelNameTaskID model.LabelName
	// dedicatedLabelNameTaskHostname is the dedicated label to use when filtering metrics on a hostname basis.
	// Storing this quick access instead of performing another O(n) search through labels.
	dedicatedLabelNameTaskHostname model.LabelName
	// Time duration for range query.
	rangeTimeUnit query.TimeUnit
	rangeQty      uint
}

func (s *TaskRankCpuSharesStrategy) Init() {
	s.rangeTimeUnit = query.None // By default cpu-shares ranking does not require a range query.
	s.rangeQty = 0
}

func (s *TaskRankCpuSharesStrategy) SetPrometheusScrapeInterval(_ time.Duration) {}

// SetTaskRanksReceiver sets the receiver of the results of task ranking.
func (s *TaskRankCpuSharesStrategy) SetTaskRanksReceiver(receiver TaskRanksReceiver) {
	s.receiver = receiver
}

// Execute the strategy using the provided data.
func (s *TaskRankCpuSharesStrategy) Execute(data model.Value) {
	valueT := data.Type()
	var matrix model.Matrix
	var vector model.Vector
	// Safety check to make sure that we cast to matrix/vector based on valueT.
	// Note, however, that as the strategy decides the metric and the range for fetching
	// data, it can assume the value type.
	// For example, if a range is provided, then the value type would be a matrix.
	// If no range is provided, then the value type would be a vector.
	switch valueT {
	case model.ValMatrix:
		matrix = data.(model.Matrix)
	case model.ValVector:
		vector = data.(model.Vector)
	default:
		// invalid value type.
		// TODO do not ignore this. maybe log it?
	}

	// Initializing tasks to rank.
	var tasks = make(entities.RankedTasks)
	addEntryForTask := func(metric model.Metric, weight float64) {
		// Fetching hostname and adding entry for host and task.
		if hostname, ok := metric[s.dedicatedLabelNameTaskHostname]; ok {
			// Adding entry for host if needed.
			if _, ok := tasks[entities.Hostname(hostname)]; !ok {
				tasks[entities.Hostname(hostname)] = make([]entities.Task, 0)
			}

			// Fetching task id and adding entry for task.
			if taskID, ok := metric[s.dedicatedLabelNameTaskID]; ok {
				tasks[entities.Hostname(hostname)] = append(tasks[entities.Hostname(hostname)],
					entities.Task{
						Metric:   metric,
						ID:       string(taskID),
						Hostname: string(hostname),
						Weight:   weight,
					})
			} else {
				// SHOULD NOT BE HERE.
			}
		} else {
			// SHOULD NOT BE HERE.
		}
	}

	if matrix != nil {
		for _, sampleStream := range matrix {
			// As cpu shares allocated to a container can be updated for docker containers,
			// taking the average of allocated cpu shares.
			addEntryForTask(sampleStream.Metric, s.avgCpuShare(sampleStream.Values))
		}
	} else if vector != nil {
		for _, sample := range vector {
			addEntryForTask(sample.Metric, float64(sample.Value))
		}
	}

	// Sorting colocated tasks in non-increasing order of cpu shares.
	for _, colocatedTasks := range tasks {
		sort.SliceStable(colocatedTasks, func(i, j int) bool {
			return colocatedTasks[i].Weight > colocatedTasks[j].Weight
		})
	}

	// Submitting the ranked tasks to the receiver.
	logger.WithFields(logrus.Fields{
		topic.TaskRankingStrategy.String(): "cpushares",
		topic.TaskRankingResult.String():   tasks,
	}).Log(logrus.InfoLevel, "strategy executed")
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
// This strategy mandates that a dedicated label be provided for filtering metrics based on TaskID and Hostname.
func (s *TaskRankCpuSharesStrategy) SetLabelMatchers(labelMatchers []*query.LabelMatcher) error {
	if len(labelMatchers) == 0 {
		return errors.New("no label matchers provided")
	}
	var foundDedicatedLabelMatcherTaskID bool
	var foundDedicatedLabelMatcherTaskHostname bool
	for _, l := range labelMatchers {
		if !l.Type.IsValid() {
			return errors.New("invalid label matcher type")
		} else if l.Type == query.TaskID {
			foundDedicatedLabelMatcherTaskID = true
			s.dedicatedLabelNameTaskID = model.LabelName(l.Label)
		} else if l.Type == query.TaskHostname {
			foundDedicatedLabelMatcherTaskHostname = true
			s.dedicatedLabelNameTaskHostname = model.LabelName(l.Label)
		}
	}

	if !foundDedicatedLabelMatcherTaskID {
		return errors.New("no dedicated task ID label matcher found")
	} else if !foundDedicatedLabelMatcherTaskHostname {
		return errors.New("no dedicated task hostname label matcher found")
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
	return s.rangeTimeUnit, s.rangeQty
}

// SetRange sets the time duration for the range query.
func (s *TaskRankCpuSharesStrategy) SetRange(timeUnit query.TimeUnit, qty uint) {
	s.rangeTimeUnit = timeUnit
	s.rangeQty = qty
}
