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
package taskranker

import (
	"fmt"
	"github.com/pradykaushik/task-ranker/datafetcher"
	"github.com/pradykaushik/task-ranker/datafetcher/prometheus"
	"github.com/pradykaushik/task-ranker/entities"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type dummyTaskRanksReceiver struct {
	rankedTasks entities.RankedTasks
}

func (r *dummyTaskRanksReceiver) Receive(rankedTasks entities.RankedTasks) {
	r.rankedTasks = rankedTasks
	fmt.Println(rankedTasks)
}

var dummyReceiver *dummyTaskRanksReceiver

func initTaskRanker(strategy string) (*TaskRanker, error) {
	var prometheusDataFetcher datafetcher.Interface
	var err error
	var tRanker *TaskRanker

	prometheusDataFetcher, err = prometheus.NewDataFetcher(
		prometheus.WithPrometheusEndpoint("http://localhost:9090"))
	if err != nil {
		return nil, err
	}

	dummyReceiver = new(dummyTaskRanksReceiver)
	tRanker, err = New(
		WithDataFetcher(prometheusDataFetcher),
		WithSchedule("?/5 * * * * *"),
		WithPrometheusScrapeInterval(1*time.Second),
		WithStrategy(strategy, []*query.LabelMatcher{
			{Type: query.TaskID, Label: "container_label_task_id", Operator: query.EqualRegex, Value: "hello_.*"},
			{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"},
		}, dummyReceiver))

	return tRanker, err
}

func initTaskRankerOptions(strategy string) (*TaskRanker, error) {
	var prometheusDataFetcher datafetcher.Interface
	var err error
	var tRanker *TaskRanker

	prometheusDataFetcher, err = prometheus.NewDataFetcher(
		prometheus.WithPrometheusEndpoint("http://localhost:9090"))
	if err != nil {
		return nil, err
	}

	dummyReceiver = new(dummyTaskRanksReceiver)
	tRanker, err = New(
		WithDataFetcher(prometheusDataFetcher),
		WithSchedule("?/5 * * * * *"),
		WithPrometheusScrapeInterval(1*time.Second),
		WithStrategyOptions(strategy,
			strategies.WithLabelMatchers([]*query.LabelMatcher{
				{Type: query.TaskID, Label: "container_label_task_id", Operator: query.EqualRegex, Value: "hello_.*"},
				{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"}}),
			strategies.WithTaskRanksReceiver(dummyReceiver)))

	return tRanker, err

}

// Test the cpushares task ranking strategy.
func TestTaskRanker_CpuSharesRanking(t *testing.T) {
	tRanker, initErr := initTaskRanker("cpushares")
	testStrategy(t, tRanker, initErr)
}

// Test the cpuutil task ranking strategy.
func TestTaskRanker_CpuUtilRanking(t *testing.T) {
	tRanker, initErr := initTaskRankerOptions("cpuutil")
	testStrategy(t, tRanker, initErr)
}

func testStrategy(t *testing.T, tRanker *TaskRanker, initErr error) {
	assert.NoError(t, initErr)
	assert.NotNil(t, tRanker)
	tRanker.Start()
	<-time.After(13 * time.Second) // Enough time for at least one round of ranking.
	testRanked(t)
	tRanker.Stop()
}

func testRanked(t *testing.T) {
	for hostname, rankedColocatedTasks := range dummyReceiver.rankedTasks {
		assert.NotEmpty(t, hostname)
		assert.NotNil(t, rankedColocatedTasks)
		assert.NotEmpty(t, rankedColocatedTasks)

		var prevTask = rankedColocatedTasks[0]
		for i := 1; i < len(rankedColocatedTasks); i++ {
			assert.Equal(t, prevTask.Hostname, "localhost")
			assert.Contains(t, prevTask.ID, "hello_")
			assert.True(t, prevTask.Weight >= rankedColocatedTasks[i].Weight)
			prevTask = rankedColocatedTasks[i]
		}
	}
}
