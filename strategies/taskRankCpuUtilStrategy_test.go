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
	"github.com/pradykaushik/task-ranker/logger"
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
		assert.Equal(t, query.None, timeUnit)
		assert.Equal(t, uint(0), qty)
	}

	count := 5
	for count > 1 {
		s.prometheusScrapeInterval = time.Duration(rand.Int63n(10)) * time.Second
		checkRange(s)
		count--
	}
}

var elapsedTimeSeconds float64 = 0

const hostname model.LabelValue = "localhost"

// Task IDs to create mocks.
var uniqueTaskSets = map[int][]model.LabelValue{
	0: {"test_task_id_1", "test_task_id_2", "test_task_id_3"},
	1: {"test_task_id_4", "test_task_id_5", "test_task_id_6"},
	2: {"test_task_id_7", "test_task_id_8", "test_task_id_9"},
}

var availableCpus = map[int]model.LabelValue{
	0: "cpu00",
	1: "cpu01",
}

// mockCpuUtilDataAlwaysUsingAllCpus returns a mock of prometheus time series data.
// This mock is useful to test scenarios where tasks are N-level parallel (N >= #cpus) and use up all the cpus all the time.
//
// This mock has the following information.
// 1. Three tasks with ids 'test_task_id_{1..3}'.
// 2. Hostname for all tasks is localhost.
// 3. For each task, cpu usage data is provided for both cpus, 'cpu00' and 'cpu01'.
// 4. task with id 'test_task_id_1' demonstrates cpu utilization of 22.5% on each cpu.
// 5. task with id 'test_task_id_2' demonstrates cpu utilization of 30% on each cpu.
// 6. task with id 'test_task_id_3' demonstrates cpu utilization of 67.5% on each cpu.
func mockCpuUtilDataAlwaysUsingAllCpus(dedicatedLabelTaskID, dedicatedLabelTaskHost model.LabelName) (mockedCpuUtilData model.Value) {
	mockedCpuUtilData = model.Value(model.Vector{
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][0],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[0],
			},
			Value:     model.SampleValue(0.225 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][0],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[1],
			},
			Value:     model.SampleValue(0.225 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][1],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[0],
			},
			Value:     model.SampleValue(0.3 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][1],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[1],
			},
			Value:     model.SampleValue(0.3 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][2],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[0],
			},
			Value:     model.SampleValue(0.675 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][2],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[1],
			},
			Value:     model.SampleValue(0.675 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
	})
	elapsedTimeSeconds++
	return
}

// mockCpuUtilDataUsingOnlySomeCpus returns a mock of prometheus time series data.
// This mock is useful to test scenarios where tasks are using only some of the available cpus.
// In addition, this mock mimics real behavior of the OS cpu scheduler assigning threads to any available cpu.
//
// This mock has the following information.
// 1. Three tasks with ids 'test_task_id_{1..3}'.
// 2. Hostname for all tasks is localhost.
// 3. For each task, cpu usage data is provided for a subset of the two cpus, 'cpu00' and 'cpu01'.
// 4. task with id 'test_task_id_1' demonstrates total cpu utilization of 45%.
// 5. task with id 'test_task_id_2' demonstrates total cpu utilization of 60%.
// 6. task with id 'test_task_id_3' demonstrates total cpu utilization of 135%.
func mockCpuUtilDataUsingOnlySomeCpus(dedicatedLabelTaskID, dedicatedLabelTaskHost model.LabelName) (mockedCpuUtilData model.Value) {
	mockedCpuUtilData = model.Value(model.Vector{
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][0],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[0],
			},
			Value:     model.SampleValue(0.45 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][1],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[0],
			},
			Value:     model.SampleValue(0.6 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][2],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[0],
			},
			Value:     model.SampleValue(0.9 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
		{
			Metric: map[model.LabelName]model.LabelValue{
				dedicatedLabelTaskID:   uniqueTaskSets[0][2],
				dedicatedLabelTaskHost: hostname,
				"cpu":                  availableCpus[1],
			},
			Value:     model.SampleValue(0.45 * (elapsedTimeSeconds + 1)),
			Timestamp: model.Time(1000 * (elapsedTimeSeconds + 1)),
		},
	})
	elapsedTimeSeconds++
	return
}

func mockEmptyTaskSetCpuUtilData() model.Value {
	return model.Vector{}
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
	// logger is already configured in the first test in this package.

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
				// Expected sum of cpu util (%) on cpu00 and cpu01.
				Weight: 135.0,
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_2",
					"container_label_task_host": "localhost",
					"cpu":                       "cpu00",
				},
				ID:       "test_task_id_2",
				Hostname: "localhost",
				// Expected sum of cpu util (%) on cpu00 and cpu01.
				Weight: 60.0,
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_1",
					"container_label_task_host": "localhost",
					"cpu":                       "cpu00",
				},
				ID:       "test_task_id_1",
				Hostname: "localhost",
				// Expected sum of cpu util (%) on cpu00 and cpu01.
				Weight: 45.0,
			},
		},
	}

	t.Run("no data retrieved from prometheus", func(t *testing.T) {
		for i := 0; i < 5; i++ { // Just testing multiple times.
			data := mockEmptyTaskSetCpuUtilData()
			s.Execute(data)

			assert.Empty(t, receiver.rankedTasks)
		}
	})

	t.Run("tasks demonstrate constant cpu usage and use all cpus", func(t *testing.T) {
		s.Init() // re-initializing.
		elapsedTimeSeconds = 0
		receiver.rankedTasks = make(entities.RankedTasks)
		for i := 0; i < 5; i++ {
			data := mockCpuUtilDataAlwaysUsingAllCpus("container_label_task_id", "container_label_task_host")
			s.Execute(data)

			if i == 0 {
				// No ranked tasks yet as we only have one second of data.
				assert.Empty(t, receiver.rankedTasks)
				continue
			}

			assert.Equal(t, len(expectedRankedTasks), len(receiver.rankedTasks))

			_, ok := expectedRankedTasks["localhost"]
			_, localhostIsInRankedTasks := receiver.rankedTasks["localhost"]
			assert.True(t, ok == localhostIsInRankedTasks)

			assert.ElementsMatch(t, expectedRankedTasks["localhost"], receiver.rankedTasks["localhost"])
		}
	})

	t.Run("tasks demonstrate varying cpu usage and do not run on all cpus", func(t *testing.T) {
		s.Init() // re-initializing.
		elapsedTimeSeconds = 0
		receiver.rankedTasks = make(entities.RankedTasks)
		for i := 0; i < 5; i++ {
			data := mockCpuUtilDataUsingOnlySomeCpus("container_label_task_id", "container_label_task_host")
			s.Execute(data)

			if i == 0 {
				// No ranked tasks yet as we only have one second of data.
				assert.Empty(t, receiver.rankedTasks)
				continue
			}

			assert.Equal(t, len(expectedRankedTasks), len(receiver.rankedTasks))

			_, ok := expectedRankedTasks["localhost"]
			_, localhostIsInRankedTasks := receiver.rankedTasks["localhost"]
			assert.True(t, ok == localhostIsInRankedTasks)

			assert.ElementsMatch(t, expectedRankedTasks["localhost"], receiver.rankedTasks["localhost"])
		}
	})

	t.Run("cpu usage data received for different set of tasks and on a subset of cpus", func(t *testing.T) {
		s.Init() // re-initializing.
		receiver.rankedTasks = make(entities.RankedTasks)
		// ROUND 1
		data := model.Vector{
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][0],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[0],
				},
				Value:     model.SampleValue(0.45),
				Timestamp: model.Time(1000),
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][1],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[0],
				},
				Value:     model.SampleValue(0.30),
				Timestamp: model.Time(1000),
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][2],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[0],
				},
				Value:     model.SampleValue(0.675),
				Timestamp: model.Time(1000),
			},
		}

		s.Execute(data)
		// No ranked tasks yet as we only have one second of data.
		assert.Empty(t, receiver.rankedTasks)

		// ROUND 2
		t.Log("changing the cpu for which cpu usage information is received for a task")
		data = model.Vector{
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][0],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[1],
				},
				Value:     model.SampleValue(0.90),
				Timestamp: model.Time(2000),
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][1],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[0],
				},
				Value:     model.SampleValue(0.30),
				Timestamp: model.Time(2000),
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][1],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[1],
				},
				Value:     model.SampleValue(0.60),
				Timestamp: model.Time(2000),
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][2],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[0],
				},
				Value:     model.SampleValue(1.35),
				Timestamp: model.Time(2000),
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   uniqueTaskSets[0][2],
					"container_label_task_host": hostname,
					"cpu":                       availableCpus[1],
				},
				Value:     model.SampleValue(0.675),
				Timestamp: model.Time(2000),
			},
		}

		expectedRankedTasks = map[entities.Hostname][]entities.Task{
			"localhost": {
				{
					Metric: map[model.LabelName]model.LabelValue{
						"container_label_task_id":   "test_task_id_3",
						"container_label_task_host": "localhost",
						"cpu":                       "cpu00",
					},
					ID:       "test_task_id_3",
					Hostname: "localhost",
					// Expected sum of cpu util (%) on cpu00 and cpu01.
					Weight: 135.0,
				},
				{
					Metric: map[model.LabelName]model.LabelValue{
						"container_label_task_id":   "test_task_id_2",
						"container_label_task_host": "localhost",
						"cpu":                       "cpu00",
					},
					ID:       "test_task_id_2",
					Hostname: "localhost",
					// Expected sum of cpu util (%) on cpu00 and cpu01.
					Weight: 60.0,
				},
				{
					Metric: map[model.LabelName]model.LabelValue{
						"container_label_task_id":   "test_task_id_1",
						"container_label_task_host": "localhost",
						"cpu":                       "cpu00",
					},
					ID:       "test_task_id_1",
					Hostname: "localhost",
					// Expected sum of cpu util (%) on cpu00 and cpu01.
					Weight: 45.0,
				},
			},
		}

		s.Execute(data)

		assert.Equal(t, len(expectedRankedTasks), len(receiver.rankedTasks))

		_, ok := expectedRankedTasks["localhost"]
		_, localhostIsInRankedTasks := receiver.rankedTasks["localhost"]
		assert.True(t, ok == localhostIsInRankedTasks)

		assert.ElementsMatch(t, expectedRankedTasks["localhost"], receiver.rankedTasks["localhost"])
	})

	assert.Nil(t, logger.Done())
}
