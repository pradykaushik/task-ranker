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
	"github.com/pradykaushik/task-ranker/entities"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

// cpuUtilRanksReceiver is a receiver of the results of executing the cpuutil task ranking strategy.
type cpuUtilRanksReceiver struct {
	rankedTasks entities.RankedTasks
}

func (r *cpuUtilRanksReceiver) Receive(rankedTasks entities.RankedTasks) {
	r.rankedTasks = rankedTasks
}

func initCpuUtilStrategy() *TaskRankCpuUtilStrategy {
	s := &TaskRankCpuUtilStrategy{}
	s.Init()
	return s
}

func TestTaskRankCpuUtilStrategy_SetTaskRanksReceiver(t *testing.T) {
	s := initCpuUtilStrategy()
	s.SetTaskRanksReceiver(&cpuUtilRanksReceiver{})
	assert.NotNil(t, s.receiver)
}

func TestTaskRankCpuUtilStrategy_GetMetric(t *testing.T) {
	s := initCpuUtilStrategy()
	assert.Equal(t, "container_cpu_usage_seconds_total", s.GetMetric())
}

func TestTaskRankCpuUtilStrategy_SetLabelMatchers(t *testing.T) {
	s := initCpuUtilStrategy()
	err := s.SetLabelMatchers([]*query.LabelMatcher{
		{Type: query.TaskID, Label: "test_label_1", Operator: query.NotEqual, Value: ""},
		{Type: query.TaskHostname, Label: "test_label_2", Operator: query.Equal, Value: "localhost"},
	})

	assert.NoError(t, err)
	assert.ElementsMatch(t, []*query.LabelMatcher{
		{Type: query.TaskID, Label: "test_label_1", Operator: query.NotEqual, Value: ""},
		{Type: query.TaskHostname, Label: "test_label_2", Operator: query.Equal, Value: "localhost"},
	}, s.GetLabelMatchers())
}

func TestTaskRankCpuUtilStrategy_GetRange(t *testing.T) {
	s := &TaskRankCpuUtilStrategy{prometheusScrapeInterval: 1 * time.Second}
	s.Init()

	checkRange := func(strategy *TaskRankCpuUtilStrategy) {
		timeUnit, qty := strategy.GetRange()
		assert.Equal(t, query.Seconds, timeUnit)
		assert.Equal(t, uint(5), qty)
	}

	count := 5
	for count > 1 {
		s.prometheusScrapeInterval = time.Duration(rand.Int63n(10)) * time.Second
		checkRange(s)
		count--
	}
}

// mockCpuUtilData returns a mock of prometheus time series data.
// This mock has the following information.
// 1. Three tasks with ids 'test_task_id_{1..3}'.
// 2. Hostname for all tasks is localhost.
// 3. For each task, cpu usage data is provided for two cpus, 'cpu00' and 'cpu01'.
// 4. task with id 'test_task_id_1' shows cpu utilization of 22.5% on each cpu.
// 5. task with id 'test_task_id_2' shows cpu utilization of 30% on each cpu.
// 6. task with id 'test_task_id_3' shows cpu utilization of 67.5% on each cpu.
func mockCpuUtilData(dedicatedLabelTaskID, dedicatedLabelTaskHost model.LabelName) model.Value {
	now := time.Now()
	return model.Value(model.Matrix{
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   "test_task_id_1",
				dedicatedLabelTaskHost: "localhost",
				"cpu":                  "cpu00",
			},
			Values: []model.SamplePair{
				{Timestamp: model.Time(now.Unix()), Value: 0.5},
				{Timestamp: model.Time(now.Add(1 * time.Second).Unix()), Value: 0.8},
				{Timestamp: model.Time(now.Add(2 * time.Second).Unix()), Value: 1.1},
				{Timestamp: model.Time(now.Add(3 * time.Second).Unix()), Value: 1.3},
				{Timestamp: model.Time(now.Add(4 * time.Second).Unix()), Value: 1.4},
			},
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   "test_task_id_1",
				dedicatedLabelTaskHost: "localhost",
				"cpu":                  "cpu01",
			},
			Values: []model.SamplePair{
				{Timestamp: model.Time(now.Unix()), Value: 0.5},
				{Timestamp: model.Time(now.Add(1 * time.Second).Unix()), Value: 0.8},
				{Timestamp: model.Time(now.Add(2 * time.Second).Unix()), Value: 1.1},
				{Timestamp: model.Time(now.Add(3 * time.Second).Unix()), Value: 1.3},
				{Timestamp: model.Time(now.Add(4 * time.Second).Unix()), Value: 1.4},
			},
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   "test_task_id_2",
				dedicatedLabelTaskHost: "localhost",
				"cpu":                  "cpu00",
			},
			Values: []model.SamplePair{
				{Timestamp: model.Time(now.Unix()), Value: 0.5},
				{Timestamp: model.Time(now.Add(1 * time.Second).Unix()), Value: 0.9},
				{Timestamp: model.Time(now.Add(2 * time.Second).Unix()), Value: 1.3},
				{Timestamp: model.Time(now.Add(3 * time.Second).Unix()), Value: 1.5},
				{Timestamp: model.Time(now.Add(4 * time.Second).Unix()), Value: 1.7},
			},
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   "test_task_id_2",
				dedicatedLabelTaskHost: "localhost",
				"cpu":                  "cpu01",
			},
			Values: []model.SamplePair{
				{Timestamp: model.Time(now.Unix()), Value: 0.5},
				{Timestamp: model.Time(now.Add(1 * time.Second).Unix()), Value: 0.9},
				{Timestamp: model.Time(now.Add(2 * time.Second).Unix()), Value: 1.3},
				{Timestamp: model.Time(now.Add(3 * time.Second).Unix()), Value: 1.5},
				{Timestamp: model.Time(now.Add(4 * time.Second).Unix()), Value: 1.7},
			},
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   "test_task_id_3",
				dedicatedLabelTaskHost: "localhost",
				"cpu":                  "cpu00",
			},
			Values: []model.SamplePair{
				{Timestamp: model.Time(now.Unix()), Value: 0.3},
				{Timestamp: model.Time(now.Add(1 * time.Second).Unix()), Value: 0.9},
				{Timestamp: model.Time(now.Add(2 * time.Second).Unix()), Value: 1.6},
				{Timestamp: model.Time(now.Add(3 * time.Second).Unix()), Value: 2.5},
				{Timestamp: model.Time(now.Add(4 * time.Second).Unix()), Value: 3.0},
			},
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   "test_task_id_3",
				dedicatedLabelTaskHost: "localhost",
				"cpu":                  "cpu01",
			},
			Values: []model.SamplePair{
				{Timestamp: model.Time(now.Unix()), Value: 0.3},
				{Timestamp: model.Time(now.Add(1 * time.Second).Unix()), Value: 0.9},
				{Timestamp: model.Time(now.Add(2 * time.Second).Unix()), Value: 1.6},
				{Timestamp: model.Time(now.Add(3 * time.Second).Unix()), Value: 2.5},
				{Timestamp: model.Time(now.Add(4 * time.Second).Unix()), Value: 3.0},
			},
		},
	})
}

func TestTaskRankCpuUtilStrategy_Execute(t *testing.T) {
	receiver := &cpuUtilRanksReceiver{}
	s := &TaskRankCpuUtilStrategy{
		receiver: receiver,
		labels: []*query.LabelMatcher{
			{Type: query.TaskID, Label: "container_label_task_id", Operator: query.NotEqual, Value: ""},
			{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"},
		},
		dedicatedLabelNameTaskID:       model.LabelName("container_label_task_id"),
		dedicatedLabelNameTaskHostname: model.LabelName("container_label_task_host"),
		prometheusScrapeInterval:       1 * time.Second,
	}
	s.Init()

	data := mockCpuUtilData("container_label_task_id", "container_label_task_host")
	s.Execute(data)

	expectedRankedTasks := map[entities.Hostname][]entities.Task{
		"localhost": {
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_3",
					"container_label_task_host": "localhost",
					"cpu":                       "cpu00",
				},
				ID:       "test_task_id_3",
				Hostname: "localhost",
				Weight:   135.0, // sum of cpu util (%) on cpu00 and cpu01.
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_2",
					"container_label_task_host": "localhost",
					"cpu":                       "cpu00",
				},
				ID:       "test_task_id_2",
				Hostname: "localhost",
				Weight:   60.0, // sum of cpu util (%) on cpu00 and cpu01.
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_1",
					"container_label_task_host": "localhost",
					"cpu":                       "cpu00",
				},
				ID:       "test_task_id_1",
				Hostname: "localhost",
				Weight:   45.0, // sum of cpu util (%) on cpu00 and cpu01.
			},
		},
	}

	assert.Equal(t, len(expectedRankedTasks), len(receiver.rankedTasks))

	_, ok := expectedRankedTasks["localhost"]
	_, localhostIsInRankedTasks := receiver.rankedTasks["localhost"]
	assert.True(t, ok == localhostIsInRankedTasks)

	assert.ElementsMatch(t, expectedRankedTasks["localhost"], receiver.rankedTasks["localhost"])
}
