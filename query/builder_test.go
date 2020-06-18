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

package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testBuilder *Builder

func TestMain(m *testing.M) {
	testBuilder = GetBuilder(
		WithMetric("test_metric"),
		WithLabelMatchers(
			&LabelMatcher{
				Label:    "test_label1",
				Operator: Equal,
				Value:    "test_value1",
			},
			&LabelMatcher{
				Label:    "test_label2",
				Operator: Equal,
				Value:    "test_value2",
			}),
		WithRange(Seconds, 5))

	m.Run()
}

func TestGetBuilder(t *testing.T) {
	assert.NotNil(t, testBuilder)
	assert.Equal(t, "test_metric", testBuilder.metric)
	assert.Len(t, testBuilder.labelMatchers, 2)
	assert.ObjectsAreEqualValues(&LabelMatcher{
		Label:    "test_label1",
		Operator: Equal,
		Value:    "test_value1",
	}, testBuilder.labelMatchers[0])
	assert.ObjectsAreEqualValues(&LabelMatcher{
		Label:    "test_label2",
		Operator: Equal,
		Value:    "test_value2",
	}, testBuilder.labelMatchers[1])
	assert.Equal(t, Seconds, testBuilder.timeUnit)
	assert.Equal(t, uint(5), testBuilder.timeDuration)
}

func TestBuilder_BuildQuery(t *testing.T) {
	const expectedQueryString = "test_metric{test_label1=\"test_value1\",test_label2=\"test_value2\"}[5s]"
	assert.Equal(t, expectedQueryString, testBuilder.BuildQuery())
}
