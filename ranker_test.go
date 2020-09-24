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
	"github.com/pradykaushik/task-ranker/datafetcher"
	"github.com/pradykaushik/task-ranker/datafetcher/prometheus"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	test := func(tRanker *TaskRanker, initErr error) {
		assert.NoError(t, initErr)
		assert.NotNil(t, tRanker)
		assert.NotNil(t, tRanker.DataFetcher)
		assert.NotNil(t, tRanker.Strategy)
		assert.Equal(t, "http://localhost:9090",
			tRanker.DataFetcher.(*prometheus.DataFetcher).GetEndpoint())
		assert.ElementsMatch(t, []*query.LabelMatcher{
			{Type: query.TaskID, Label: "container_label_task_id", Operator: query.EqualRegex, Value: "hello_.*"},
			{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"},
		}, tRanker.Strategy.(*strategies.TaskRankCpuSharesStrategy).GetLabelMatchers())
		parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		tRankerSchedule, err := parser.Parse("?/5 * * * * *")
		assert.NoError(t, err)
		assert.Equal(t, tRankerSchedule, tRanker.Schedule)
	}

	t.Run("using WithStrategy", func(t *testing.T) {
		tRanker, initErr := initTaskRanker("cpushares")
		test(tRanker, initErr)
	})

	t.Run("using WithStrategyOptions", func(t *testing.T) {
		tRanker, initErr := initTaskRankerOptions("cpushares")
		test(tRanker, initErr)
		timeUnit, qty := tRanker.Strategy.GetRange()
		assert.Equal(t, query.None, timeUnit)
		assert.Equal(t, uint(0), qty)
	})
}

func TestNew_InvalidSchedule(t *testing.T) {
	var prometheusDataFetcher datafetcher.Interface
	var err error
	var tRanker *TaskRanker

	prometheusDataFetcher, err = prometheus.NewDataFetcher(
		prometheus.WithPrometheusEndpoint("http://localhost:9090"))
	assert.NoError(t, err, "failed to instantiate data fetcher")

	dummyReceiver = new(dummyTaskRanksReceiver)
	t.Run("schedule (in seconds) = 0", func(t *testing.T) {
		// Setting the task ranker schedule to be 0.
		tRanker, err = New(
			WithDataFetcher(prometheusDataFetcher),
			WithSchedule("?/0 * * * * *"),
			WithPrometheusScrapeInterval(1*time.Second),
			WithStrategyOptions("cpushares",
				strategies.WithLabelMatchers([]*query.LabelMatcher{
					{Type: query.TaskID, Label: "container_label_task_id", Operator: query.EqualRegex, Value: "hello_.*"},
					{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"}}),
				strategies.WithTaskRanksReceiver(dummyReceiver)))
		assert.Error(t, err, "task ranker schedule validation failed")
		assert.Nil(t, tRanker, "task ranker instantiated with invalid schedule")
	})

	t.Run("schedule (in seconds) = prometheus scrape interval = 1", func(t *testing.T) {
		// Setting the task ranker schedule to every 1 second.
		// Setting the prometheus scrape interval to 1 second.
		tRanker, err = New(
			WithDataFetcher(prometheusDataFetcher),
			WithSchedule("?/1 * * * * *"),
			WithPrometheusScrapeInterval(1*time.Second),
			WithStrategyOptions("cpushares",
				strategies.WithLabelMatchers([]*query.LabelMatcher{
					{Type: query.TaskID, Label: "container_label_task_id", Operator: query.EqualRegex, Value: "hello_.*"},
					{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"}}),
				strategies.WithTaskRanksReceiver(dummyReceiver)))
		assert.NoError(t, err, "task ranker schedule validation failed")
		assert.NotNil(t, tRanker, "task ranker not instantiated in-spite of valid configuration")
	})

	t.Run("schedule (in seconds) < prometheus scrape interval", func(t *testing.T) {
		// Setting the task ranker schedule to every 3 seconds (< 5 seconds).
		tRanker, err = New(
			WithDataFetcher(prometheusDataFetcher),
			WithSchedule("?/3 * * * * *"),
			WithPrometheusScrapeInterval(5*time.Second),
			WithStrategyOptions("cpushares",
				strategies.WithLabelMatchers([]*query.LabelMatcher{
					{Type: query.TaskID, Label: "container_label_task_id", Operator: query.EqualRegex, Value: "hello_.*"},
					{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"}}),
				strategies.WithTaskRanksReceiver(dummyReceiver)))
		assert.Error(t, err, "task ranker schedule validation failed")
		assert.Nil(t, tRanker, "task ranker instantiated with invalid schedule")
	})

	t.Run("task ranker schedule (in seconds) is not a positive multiple of prometheus scrape interval", func(t *testing.T) {
		// Setting the task ranker schedule to every 7 seconds (not a positive multiple of 5).
		tRanker, err = New(
			WithDataFetcher(prometheusDataFetcher),
			WithSchedule("?/7 * * * * *"),
			WithPrometheusScrapeInterval(5*time.Second),
			WithStrategyOptions("cpushares",
				strategies.WithLabelMatchers([]*query.LabelMatcher{
					{Type: query.TaskID, Label: "container_label_task_id", Operator: query.EqualRegex, Value: "hello_.*"},
					{Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"}}),
				strategies.WithTaskRanksReceiver(dummyReceiver)))
		assert.Error(t, err, "task ranker schedule validation failed")
		assert.Nil(t, tRanker, "task ranker instantiated with invalid schedule")
	})
}
