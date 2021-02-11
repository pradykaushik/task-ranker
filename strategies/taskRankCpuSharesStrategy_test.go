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
	"testing"
	"time"
)

// cpuSharesRanksReceiver is a receiver of the results of executing the cpushares task ranking strategy.
type cpuSharesRanksReceiver struct {
	rankedTasks entities.RankedTasks
}

func (r *cpuSharesRanksReceiver) Receive(rankedTasks entities.RankedTasks) {
	r.rankedTasks = rankedTasks
}

func initCpusharesStrategy() *TaskRankCpuSharesStrategy {
	s := &TaskRankCpuSharesStrategy{}
	s.Init()
	return s
}

func TestTaskRankCpuSharesStrategy_SetTaskRanksReceiver(t *testing.T) {
	s := initCpusharesStrategy()
	s.SetTaskRanksReceiver(&cpuSharesRanksReceiver{})
	assert.NotNil(t, s.receiver)
}

func TestTaskRankCpuSharesStrategy_GetMetric(t *testing.T) {
	s := initCpusharesStrategy()
	assert.Equal(t, "container_spec_cpu_shares", s.GetMetrics()[0])
}

func TestTaskRankCpuSharesStrategy_SetLabelMatchers(t *testing.T) {
	s := initCpusharesStrategy()
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

func TestTaskRankCpuSharesStrategy_GetRange(t *testing.T) {
	s := initCpusharesStrategy()
	timeUnit, qty := s.GetRange()
	assert.Equal(t, query.None, timeUnit)
	assert.Equal(t, uint(0), qty)
}

// mockCpuSharesData returns a mock of prometheus time series data.
// This mock has the following information.
// 1. Three tasks with ids 'test_task_id_{1..3}'.
// 2. Hostname for all tasks is localhost.
// 3. task with id 'test_task_id_1' is allocated a 1024 cpu shares.
// 4. task with id 'test_task_id_2' is allocated a 2048 cpu shares.
// 5. task with id 'test_task_id_3' is allocated a 3072 cpu shares.
func mockCpuSharesData(dedicatedLabelTaskID, dedicatedLabelTaskHost model.LabelName) model.Value {
	now := time.Now()
	return model.Matrix{
		getMockDataRange(dedicatedLabelTaskID, dedicatedLabelTaskHost, uniqueTaskSets[0][0],
			hostname, model.SamplePair{Timestamp: model.Time(now.Second()), Value: 1024.0}),
		getMockDataRange(dedicatedLabelTaskID, dedicatedLabelTaskHost, uniqueTaskSets[0][1],
			hostname, model.SamplePair{Timestamp: model.Time(now.Second()), Value: 2048.0}),
		getMockDataRange(dedicatedLabelTaskID, dedicatedLabelTaskHost, uniqueTaskSets[0][2],
			hostname, model.SamplePair{Timestamp: model.Time(now.Second()), Value: 3072.0}),
	}
}

func TestTaskRankCpuSharesStrategy_Execute(t *testing.T) {
	receiver := &cpuSharesRanksReceiver{}
	s := &TaskRankCpuSharesStrategy{
		receiver: receiver,
		labels: []*query.LabelMatcher{
			{Type: query.TaskID, Label: "container_label_task_id", Operator: query.NotEqual, Value: ""},
			{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"},
		},
		dedicatedLabelNameTaskID:       model.LabelName("container_label_task_id"),
		dedicatedLabelNameTaskHostname: model.LabelName("container_label_task_host"),
	}
	s.Init()

	data := mockCpuSharesData("container_label_task_id", "container_label_task_host")
	s.Execute(data)

	expectedRankedTasks := map[entities.Hostname][]entities.Task{
		"localhost": {
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_3",
					"container_label_task_host": "localhost",
				},
				ID:       "test_task_id_3",
				Hostname: "localhost",
				Weight:   3072.0,
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_2",
					"container_label_task_host": "localhost",
				},
				ID:       "test_task_id_2",
				Hostname: "localhost",
				Weight:   2048.0,
			},
			{
				Metric: map[model.LabelName]model.LabelValue{
					"container_label_task_id":   "test_task_id_1",
					"container_label_task_host": "localhost",
				},
				ID:       "test_task_id_1",
				Hostname: "localhost",
				Weight:   1024.0,
			},
		},
	}

	assert.Equal(t, len(expectedRankedTasks), len(receiver.rankedTasks))

	_, ok := expectedRankedTasks["localhost"]
	_, localhostIsInRankedTasks := receiver.rankedTasks["localhost"]
	assert.True(t, ok == localhostIsInRankedTasks)

	assert.ElementsMatch(t, expectedRankedTasks["localhost"], receiver.rankedTasks["localhost"])
}
