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
	"fmt"
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/entities"
	"github.com/pradykaushik/task-ranker/logger"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"math"
	"sort"
	"time"
)

// TaskRankCpuUtilStrategy is a task ranking strategy that ranks the tasks
// in non-increasing order based on their cpu utilization (%).
//
// For example, if Prometheus scrapes metrics every 1s, then this strategy would rank tasks based
// on their cpu utilization in the past second.
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
	// previousTotalCpuUsage stores the sum of cumulative cpu usage seconds for each running task.
	previousTotalCpuUsage map[string]map[string]*taskCpuUsageInfo
	// Time duration for range query.
	// Note that there is a caveat in using range queries to retrieve cpu time for containers.
	// If the tasks are pinned, then using a range > 1s works as we would always get the necessary data points for each cpu (thread).
	// On the other hand, if the tasks are not pinned, then there is no guarantee that the necessary number of data points be available
	// as the cpu scheduler can preempt and re-schedule the task on any available cpu.
	// Therefore, to avoid confusion, this strategy does not use the range query.
	rangeTimeUnit query.TimeUnit
	rangeQty      uint
}

type taskCpuUsageInfo struct {
	task      *entities.Task
	dataPoint *cpuUsageDataPoint
}

type cpuUsageDataPoint struct {
	totalCumulativeCpuUsage float64
	timestamp               model.Time
}

func (d cpuUsageDataPoint) String() string {
	return fmt.Sprintf("cumulativeCpuUsage[%f], Timestamp[%d]", d.totalCumulativeCpuUsage, d.timestamp)
}

func (c taskCpuUsageInfo) String() string {
	prev := "nil"
	if c.dataPoint != nil {
		prev = fmt.Sprintf("%s", c.dataPoint)
	}

	return fmt.Sprintf("Weight[%f], Prev[%s]", c.task.Weight, prev)
}

func (s *TaskRankCpuUtilStrategy) Init() {
	s.rangeTimeUnit = query.None
	s.rangeQty = 0
	s.previousTotalCpuUsage = make(map[string]map[string]*taskCpuUsageInfo)
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
	var vector model.Vector
	// Safety check to make sure that we cast to vector only if value type is vector.
	// Note, however, that as the strategy decides the metric and the range for fetching data,
	// it can assume the value type.
	// For example, if a range is provided, then the value type would be a matrix.
	// If no range is provided, the value type would be a vector.
	switch valueT {
	case model.ValVector:
		vector = data.(model.Vector)
	default:
		// invalid value type.
		// TODO do not ignore this. maybe log it?
	}

	var ok bool
	// nowTotalCpuUsage stores the total cumulative cpu usage for each running task.
	var nowTotalCpuUsage = make(map[string]map[string]*cpuUsageDataPoint)

	// Parse Prometheus metrics.
	// TODO (pradykaushik) make efficient and parallelize parsing of Prometheus metrics.
	for _, sample := range vector {
		var hostname model.LabelValue
		if hostname, ok = sample.Metric[s.dedicatedLabelNameTaskHostname]; ok {
			if _, ok = s.previousTotalCpuUsage[string(hostname)]; !ok {
				// First time fetching metrics from this host.
				s.previousTotalCpuUsage[string(hostname)] = make(map[string]*taskCpuUsageInfo)
			}

			if _, ok = nowTotalCpuUsage[string(hostname)]; !ok {
				// Creating entry to record current total cumulative cpu usage for colocated tasks.
				nowTotalCpuUsage[string(hostname)] = make(map[string]*cpuUsageDataPoint)
			}

			var taskID model.LabelValue
			if taskID, ok = sample.Metric[s.dedicatedLabelNameTaskID]; ok {
				if _, ok := s.previousTotalCpuUsage[string(hostname)][string(taskID)]; !ok {
					// First time fetching metrics for task. Recording taskID and hostname to help consolidation.
					taskTotalCpuUsage := &taskCpuUsageInfo{
						task: &entities.Task{
							Metric:   sample.Metric,
							Weight:   0.0,
							ID:       string(taskID),
							Hostname: string(hostname),
						},
						dataPoint: nil,
					}

					s.previousTotalCpuUsage[string(hostname)][string(taskID)] = taskTotalCpuUsage
				}

				if _, ok := nowTotalCpuUsage[string(hostname)][string(taskID)]; !ok {
					// Recording cumulative cpu usage seconds for task on cpu.
					nowTotalCpuUsage[string(hostname)][string(taskID)] = &cpuUsageDataPoint{
						totalCumulativeCpuUsage: float64(sample.Value),
						timestamp:               sample.Timestamp,
					}
				} else {
					// Adding cumulative cpu usage seconds for task on a cpu.
					nowTotalCpuUsage[string(hostname)][string(taskID)].totalCumulativeCpuUsage += float64(sample.Value)
				}
			} else {
				// Either taskID not exported along with the metrics, or
				// Task ID dedicated label provided is incorrect.
				// TODO do not ignore this. We should log this instead.
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
	for hostname, colocatedTasksCpuUsageInfo := range s.previousTotalCpuUsage {
		for taskID, prevTotalCpuUsage := range colocatedTasksCpuUsageInfo {
			if prevTotalCpuUsage.dataPoint == nil {
				prevTotalCpuUsage.dataPoint = new(cpuUsageDataPoint)
			} else {
				// Calculating the cpu utilization of this task.
				prevTotalCpuUsage.task.Weight = s.round(s.cpuUtil(
					prevTotalCpuUsage.dataPoint.totalCumulativeCpuUsage,
					prevTotalCpuUsage.dataPoint.timestamp,
					nowTotalCpuUsage[hostname][taskID].totalCumulativeCpuUsage,
					nowTotalCpuUsage[hostname][taskID].timestamp,
				))
				rankedTasks[entities.Hostname(hostname)] = append(rankedTasks[entities.Hostname(hostname)], *prevTotalCpuUsage.task)
			}
			// Saving current total cumulative cpu usage seconds to calculate cpu utilization in the next interval.
			prevTotalCpuUsage.dataPoint.totalCumulativeCpuUsage = nowTotalCpuUsage[hostname][taskID].totalCumulativeCpuUsage
			prevTotalCpuUsage.dataPoint.timestamp = nowTotalCpuUsage[hostname][taskID].timestamp
		}

		// Sorting colocated tasks.
		sort.SliceStable(rankedTasks[entities.Hostname(hostname)], func(i, j int) bool {
			return rankedTasks[entities.Hostname(hostname)][i].Weight >=
				rankedTasks[entities.Hostname(hostname)][j].Weight
		})
	}

	// Submitting ranked tasks to the receiver.
	if len(rankedTasks) > 0 {
		logger.WithFields(logrus.Fields{
			"task_ranking_strategy": "cpuutil",
			"task_ranking_results":  rankedTasks,
		}).Log(logrus.InfoLevel, "strategy executed")
		s.receiver.Receive(rankedTasks)
	}
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
// The time difference (in seconds) of the two data points is used as the elapsed time. Note that this okay
// as the task ranker schedule (in seconds) is a multiple of the prometheus scrape interval.
func (s TaskRankCpuUtilStrategy) cpuUtil(
	prevTotalCpuUsage float64,
	prevTotalCpuUsageTimestamp model.Time,
	nowTotalCpuUsage float64,
	nowTotalCpuUsageTimestamp model.Time) float64 {

	// timestamps are in milliseconds and therefore dividing by 1000 to convert to seconds.
	timeDiffSeconds := float64(nowTotalCpuUsageTimestamp-prevTotalCpuUsageTimestamp) / 1000
	return 100.0 * ((nowTotalCpuUsage - prevTotalCpuUsage) / timeDiffSeconds)
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
}

// SetRange sets the time duration for the range query.
// As this strategy does not use range queries a call to this method results in a NO-OP.
func (s *TaskRankCpuUtilStrategy) SetRange(_ query.TimeUnit, _ uint) {}
