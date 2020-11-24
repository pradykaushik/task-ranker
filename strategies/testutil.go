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

import "github.com/prometheus/common/model"

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

// getMockDataSample returns a data sample mimicking cpu_usage_seconds information
// retrieved from prometheus for a single task.
func getMockDataSample(
	dedicatedLabelTaskID, dedicatedLabelTaskHost model.LabelName,
	taskID, hostname, cpu model.LabelValue,
	cumulativeCpuUsageSeconds float64,
	timestamp float64) *model.Sample {

	return &model.Sample{
		Metric: map[model.LabelName]model.LabelValue{
			dedicatedLabelTaskID:   taskID,
			dedicatedLabelTaskHost: hostname,
			"cpu":                  cpu,
		},
		Value:     model.SampleValue(cumulativeCpuUsageSeconds),
		Timestamp: model.Time(timestamp),
	}
}

// getMockDataRange returns a data sample stream mimicking cpu_usage_seconds information
// retrieved as a result of a range query from prometheus for a single task.
func getMockDataRange(
	dedicatedLabelTaskID, dedicatedLabelTaskHost model.LabelName,
	taskID, hostname model.LabelValue,
	values ...model.SamplePair) *model.SampleStream {

	return &model.SampleStream{
		Metric: map[model.LabelName]model.LabelValue{
			dedicatedLabelTaskID:   taskID,
			dedicatedLabelTaskHost: hostname,
		},
		Values: values,
	}
}
