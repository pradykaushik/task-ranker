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
	"github.com/pradykaushik/task-ranker/datafetcher/prometheus"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
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
		assert.Equal(t, query.Seconds, timeUnit)
		assert.Equal(t, uint(5), qty)
	})
}
