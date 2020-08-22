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
	"math"
	"sort"
	"time"
)

// TaskRankCpuUtilStrategy is a task ranking strategy that ranks the tasks
// in non-increasing order based on the cpu utilization (%) in the past 5 intervals of time.
//
// For example, if Prometheus scrapes metrics every 1s, then each time interval is 1s long.
// This strategy then would then rank tasks based on their cpu utilization in the past 5 seconds.
type TaskRankCpuUtilStrategy struct {
	// receiver of the results of task ranking.
	receiver TaskRanksReceiver
	// labels used to filter the time series data fetched from prometheus.
	labels []*query.LabelMatcher
	// dedicatedLabelNameTaskID is the dedicated label to use when filtering metrics based on task id.
	// Storing this for quick access to prevent another O(n) search through labels.
	dedicatedLabelNameTaskID model.LabelName
	// dedicatedLabelNameTaskHostname is the dedicated label to use when filtering metrics on a hostname basis.
	// Storing this for quick access to prevent another O(n) search through labels.
	dedicatedLabelNameTaskHostname model.LabelName
	// prometheusScrapeInterval corresponds to the time interval between two successive metrics scrapes.
	prometheusScrapeInterval time.Duration
	// Time duration for range query.
	rangeTimeUnit query.TimeUnit
	rangeQty      uint
}

func (s *TaskRankCpuUtilStrategy) Init() {
	// By default, rank tasks based on past 5 seconds cpu usage.
	s.rangeTimeUnit = query.Seconds
	s.rangeQty = 5
}

// SetPrometheusScrapeInterval sets the scrape interval of prometheus.
func (s *TaskRankCpuUtilStrategy) SetPrometheusScrapeInterval(prometheusScrapeInterval time.Duration) {
	s.prometheusScrapeInterval = prometheusScrapeInterval
}

// SetTaskRanksReceiver sets the receiver of the task ranking results.
func (s *TaskRankCpuUtilStrategy) SetTaskRanksReceiver(receiver TaskRanksReceiver) {
	s.receiver = receiver
}

// Execute the strategy using the provided data.
func (s *TaskRankCpuUtilStrategy) Execute(data model.Value) {
	valueT := data.Type()
	var matrix model.Matrix
	// Safety check to make sure that we cast to matrix only if value type is matrix.
	// Note, however, that as the strategy decides the metric and the range for fetching data,
	// it can assume the value type. For example, if a range is provided, then the value type would
	// be a matrix.
	switch valueT {
	case model.ValMatrix:
		matrix = data.(model.Matrix)
	default:
		// invalid value type.
		// TODO do not ignore this. maybe log it?
	}

	// cpuLabelName is the label name that can be used to fetch the cpu associated with the usage data.
	const cpuLabelName model.LabelName = "cpu"

	type hostValueT model.LabelValue
	type taskIDValueT model.LabelValue

	var ok bool
	// allTasksCpuUsageSeconds stores the cpu utilization (%) for all tasks on each cpu on each host.
	allTasksCpuUsageSeconds := make(map[hostValueT]map[taskIDValueT]*entities.Task)

	// Parse Prometheus metrics.
	// TODO (pradykaushik) make efficient and parallelize parsing of Prometheus metrics.
	for _, sampleStream := range matrix {
		// Not considering task for ranking if < 2 data points retrieved.
		if len(sampleStream.Values) < 2 {
			continue
		}

		var hostname model.LabelValue
		var colocatedTasksCpuUsageSeconds map[taskIDValueT]*entities.Task
		if hostname, ok = sampleStream.Metric[s.dedicatedLabelNameTaskHostname]; ok {
			if colocatedTasksCpuUsageSeconds, ok = allTasksCpuUsageSeconds[hostValueT(hostname)]; !ok {
				// First time fetching metrics from this host.
				colocatedTasksCpuUsageSeconds = make(map[taskIDValueT]*entities.Task)
				allTasksCpuUsageSeconds[hostValueT(hostname)] = colocatedTasksCpuUsageSeconds
			}

			var taskID model.LabelValue
			var taskCpuUsage *entities.Task
			if taskID, ok = sampleStream.Metric[s.dedicatedLabelNameTaskID]; ok {
				if taskCpuUsage, ok = colocatedTasksCpuUsageSeconds[taskIDValueT(taskID)]; !ok {
					// First time fetching metrics for task. Recording taskID and hostname to help consolidation.
					taskCpuUsage = &entities.Task{
						Metric:   sampleStream.Metric,
						Weight:   0.0,
						ID:       string(taskID),
						Hostname: string(hostname),
					}

					colocatedTasksCpuUsageSeconds[taskIDValueT(taskID)] = taskCpuUsage
				}

				if _, ok = sampleStream.Metric[cpuLabelName]; ok {
					// Calculating and recording the cpu utilization (%) of this task on this cpu.
					// Adding it to an accumulator that will later be used as the cpu utilization of
					// this task across all cpus.
					taskCpuUsage.Weight += s.cpuUtil(sampleStream.Values)
				} else {
					// CPU usage data does not correspond to a particular cpu.
					// TODO do not ignore this. We should instead log this.
				}
			}
		} else {
			// Either hostname not exported along with the metrics, or
			// Task Hostname dedicated label provided is incorrect.
			// TODO do not ignore this. We should log this instead.
		}
	}

	// Rank colocated tasks in non-increasing order based on their total cpu utilization (%)
	// on the entire host (all cpus).
	rankedTasks := make(entities.RankedTasks)
	for hostname, colocatedTasksCpuUsageSeconds := range allTasksCpuUsageSeconds {
		for _, taskCpuUsage := range colocatedTasksCpuUsageSeconds {
			taskCpuUsage.Weight = s.round(taskCpuUsage.Weight)
			rankedTasks[entities.Hostname(hostname)] = append(rankedTasks[entities.Hostname(hostname)], *taskCpuUsage)
		}

		// Sorting colocated tasks.
		sort.SliceStable(rankedTasks[entities.Hostname(hostname)], func(i, j int) bool {
			return rankedTasks[entities.Hostname(hostname)][i].Weight >=
				rankedTasks[entities.Hostname(hostname)][j].Weight
		})
	}

	// Submitting ranked tasks to the receiver.
	s.receiver.Receive(rankedTasks)
}

// cpuUtilPrecision defines the precision for cpu utilization (%) values.
// As cpu utilization can be less than 1%, the precision is set to 2.
const cpuUtilPrecision = 2

// round the cpu utilization (%) to the above specified precision.
func (s TaskRankCpuUtilStrategy) round(cpuUtil float64) float64 {
	multiplier := math.Pow10(cpuUtilPrecision)
	magnified := cpuUtil * multiplier
	return math.Round(magnified) / multiplier
}

// cpuUtil calculates and returns the cpu utilization (%) for the task.
// Elapsed time is calculated as the number of seconds between oldest and newest value by
// factoring in the prometheus scrape interval.
func (s TaskRankCpuUtilStrategy) cpuUtil(values []model.SamplePair) float64 {
	n := len(values)
	return 100.0 * ((float64(values[n-1].Value) - float64(values[0].Value)) /
		(float64(n-1) * s.prometheusScrapeInterval.Seconds()))
}

// GetMetric returns the name of the metric to query.
func (s TaskRankCpuUtilStrategy) GetMetric() string {
	// TODO convert this to constant.
	return "container_cpu_usage_seconds_total"
}

// SetLabelMatchers sets the label matchers to use when filtering data.
// This strategy mandates that a dedicated label be provided for filtering metrics based on TaskID and Hostname.
func (s *TaskRankCpuUtilStrategy) SetLabelMatchers(labelMatchers []*query.LabelMatcher) error {
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
func (s TaskRankCpuUtilStrategy) GetLabelMatchers() []*query.LabelMatcher {
	return s.labels
}

// GetRange returns the time unit and duration for how far back (in seconds) values need to be fetched.
func (s TaskRankCpuUtilStrategy) GetRange() (query.TimeUnit, uint) {
	return s.rangeTimeUnit, s.rangeQty
	// return query.Seconds, uint(5 * int(s.prometheusScrapeInterval.Seconds()))
}

// SetRange sets the time duration for the range query.
// For cpu-util ranking strategy the time duration has to be > 1s as you need two data points to calculate cpu utilization.
// If the provided time duration <= 1s, the default duration of 5 intervals of time is used, where each
// interval of time is equal to the prometheus scrape interval.
func (s *TaskRankCpuUtilStrategy) SetRange(timeUnit query.TimeUnit, qty uint) {
	if !timeUnit.IsValid() || ((timeUnit == query.Seconds) && qty <= 1) {
		s.rangeTimeUnit = query.Seconds
		s.rangeQty = uint(5 * int(s.prometheusScrapeInterval.Seconds()))
	} else {
		s.rangeTimeUnit = timeUnit
		s.rangeQty = qty
	}
}
