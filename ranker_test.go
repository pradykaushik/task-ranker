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
	"bytes"
	"fmt"
	"github.com/pradykaushik/task-ranker/datafetcher"
	"github.com/pradykaushik/task-ranker/datafetcher/prometheus"
	"github.com/pradykaushik/task-ranker/entities"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

type dummyTaskRankReceiver struct{}

func (r *dummyTaskRankReceiver) Receive(rankedTasks []entities.RankedTask) {
	var buf bytes.Buffer
	for _, t := range rankedTasks {
		buf.WriteString(fmt.Sprintf("%v\n", t.String()))
	}
	log.Println(buf.String())
}

func initTaskRanker(t *testing.T) (*TaskRanker, error) {
	var err error
	var prometheusDataFetcher datafetcher.Interface
	var tRanker *TaskRanker

	prometheusDataFetcher, err = prometheus.NewDataFetcher(
		prometheus.WithPrometheusEndpoint("http://localhost:9090"))

	assert.NoError(t, err)
	assert.NotNil(t, prometheusDataFetcher)

	tRanker, err = New(
		WithDataFetcher(prometheusDataFetcher),
		WithSchedule("?/5 * * * * *"),
		WithStrategy("cpushares", []*query.LabelMatcher{
			{Label: "label1", Operator: query.Equal},
			{Label: "label2", Operator: query.Equal},
		}, &dummyTaskRankReceiver{}))

	return tRanker, err
}

func TestNew(t *testing.T) {
	var err error
	var tRanker *TaskRanker
	tRanker, err = initTaskRanker(t)
	assert.NoError(t, err)
	assert.NotNil(t, tRanker)
	assert.NotNil(t, tRanker.DataFetcher)
	assert.NotNil(t, tRanker.Strategy)
	assert.Equal(t, "http://localhost:9090",
		tRanker.DataFetcher.(*prometheus.DataFetcher).GetEndpoint())
	assert.ElementsMatch(t, []*query.LabelMatcher{
		{Label: "label1", Operator: query.Equal},
		{Label: "label2", Operator: query.Equal},
	}, tRanker.Strategy.(*strategies.TaskRankCpuSharesStrategy).GetLabelMatchers())
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	var sched cron.Schedule
	sched, err = parser.Parse("?/5 * * * * *")
	assert.NoError(t, err)
	assert.Equal(t, sched, tRanker.Schedule)
}

// Test the cpushares task ranking strategy by fetching data from a local prometheus + cAdvisor setup.
func TestTaskRanker_Start(t *testing.T) {
	var err error
	var prometheusDataFetcher datafetcher.Interface
	var tRanker *TaskRanker

	prometheusDataFetcher, err = prometheus.NewDataFetcher(
		prometheus.WithPrometheusEndpoint("http://localhost:9090"))
	assert.NoError(t, err)
	assert.NotNil(t, prometheusDataFetcher)

	tRanker, err = New(
		WithDataFetcher(prometheusDataFetcher),
		WithSchedule("?/5 * * * * *"),
		WithStrategy("cpushares", []*query.LabelMatcher{
			{Label: "container_label_task_name", Operator: query.EqualRegex, Value: "test_task_.*"},
		}, &dummyTaskRankReceiver{}))
	assert.NoError(t, err)
	tRanker.Start()

	<-time.After(10 * time.Second)
	tRanker.Stop()
}
