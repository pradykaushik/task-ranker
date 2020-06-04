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
	"github.com/pradykaushik/task-ranker/entities"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type dummyTaskRankReceiver struct{}

func (r *dummyTaskRankReceiver) Receive(rankedTasks []entities.Task) {
	// placeholder
	log.Println("len(rankedTasks) = ", len(rankedTasks))
}

func TestNew(t *testing.T) {
	tRanker, err := New(
		WithPrometheusEndpoint("http://localhost:9090"),
		WithFilterLabelsZeroValues([]string{"label1", "label2"}),
		WithStrategy("cpushares", &dummyTaskRankReceiver{}),
		WithSchedule("?/5 * * * * *"))

	assert.NoError(t, err)
	assert.NotNil(t, tRanker)
	assert.Equal(t, "http://localhost:9090", tRanker.PrometheusEndpoint)
	assert.NotNil(t, tRanker.Strategy)
	assert.Len(t, tRanker.FilterLabels, 2)
	assert.Equal(t, "?/5 * * * * *", tRanker.Schedule)
}
