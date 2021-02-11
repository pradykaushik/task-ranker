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

var testBuilderWithRangeQuery *Builder
var testBuilderWithoutRangeQuery *Builder
var testBuilderWithMultipleMetrics *Builder

func TestNewBuilder(t *testing.T) {
	t.Run("with-range-query", func(t *testing.T) {
		testBuilderWithRangeQuery = NewBuilder(
			WithMetrics([]string{"test_metric"}),
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

		assert.NotNil(t, testBuilderWithRangeQuery)
		assert.Equal(t, "test_metric", testBuilderWithRangeQuery.metrics[0])
		assert.Len(t, testBuilderWithRangeQuery.labelMatchers, 2)
		assert.ObjectsAreEqualValues(&LabelMatcher{
			Label:    "test_label1",
			Operator: Equal,
			Value:    "test_value1",
		}, testBuilderWithRangeQuery.labelMatchers[0])
		assert.ObjectsAreEqualValues(&LabelMatcher{
			Label:    "test_label2",
			Operator: Equal,
			Value:    "test_value2",
		}, testBuilderWithRangeQuery.labelMatchers[1])
		assert.Equal(t, Seconds, testBuilderWithRangeQuery.timeUnit)
		assert.Equal(t, uint(5), testBuilderWithRangeQuery.timeDuration)
	})

	t.Run("without-range-query", func(t *testing.T) {
		testBuilderWithoutRangeQuery = NewBuilder(
			WithMetrics([]string{"test_metric"}),
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
			WithRange(None, 0))

		assert.NotNil(t, testBuilderWithoutRangeQuery)
		assert.Equal(t, "test_metric", testBuilderWithoutRangeQuery.metrics[0])
		assert.Len(t, testBuilderWithoutRangeQuery.labelMatchers, 2)
		assert.ObjectsAreEqualValues(&LabelMatcher{
			Label:    "test_label1",
			Operator: Equal,
			Value:    "test_value1",
		}, testBuilderWithoutRangeQuery.labelMatchers[0])
		assert.ObjectsAreEqualValues(&LabelMatcher{
			Label:    "test_label2",
			Operator: Equal,
			Value:    "test_value2",
		}, testBuilderWithoutRangeQuery.labelMatchers[1])
		assert.Equal(t, None, testBuilderWithoutRangeQuery.timeUnit)
		assert.Equal(t, uint(0), testBuilderWithoutRangeQuery.timeDuration)
	})

	t.Run("with-multiple-metrics", func(t *testing.T) {
		testBuilderWithMultipleMetrics = NewBuilder(
			WithMetrics([]string{"test_metric1", "test_metric2"}),
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
			WithRange(None, 0))
	})

	assert.NotNil(t, testBuilderWithMultipleMetrics)
	assert.Len(t, testBuilderWithMultipleMetrics.metrics, 2)
	assert.Equal(t, "test_metric1", testBuilderWithMultipleMetrics.metrics[0])
	assert.Equal(t, "test_metric2", testBuilderWithMultipleMetrics.metrics[1])
	assert.Len(t, testBuilderWithMultipleMetrics.labelMatchers, 2)
	assert.ObjectsAreEqualValues(&LabelMatcher{
		Label:    "test_label1",
		Operator: Equal,
		Value:    "test_value1",
	}, testBuilderWithMultipleMetrics.labelMatchers[0])
	assert.ObjectsAreEqualValues(&LabelMatcher{
		Label:    "test_label2",
		Operator: Equal,
		Value:    "test_value2",
	}, testBuilderWithMultipleMetrics.labelMatchers[1])
	assert.Equal(t, None, testBuilderWithMultipleMetrics.timeUnit)
	assert.Equal(t, uint(0), testBuilderWithMultipleMetrics.timeDuration)
}

func TestBuilder_BuildQuery(t *testing.T) {
	t.Run("with-range-query", func(t *testing.T) {
		const expectedQueryStringWithRange = "{__name__=\"test_metric\",test_label1=\"test_value1\",test_label2=\"test_value2\"}[5s]"
		assert.Equal(t, expectedQueryStringWithRange, testBuilderWithRangeQuery.BuildQuery())
	})

	t.Run("without-range-query", func(t *testing.T) {
		const expectedQueryStringWithoutRange = "{__name__=\"test_metric\",test_label1=\"test_value1\",test_label2=\"test_value2\"}"
		assert.Equal(t, expectedQueryStringWithoutRange, testBuilderWithoutRangeQuery.BuildQuery())
	})

	t.Run("with-multiple-metrics", func(t *testing.T) {
		const expectedQueryStringWithMultipleMetrics = "{__name__=~\"test_metric1|test_metric2\",test_label1=\"test_value1\",test_label2=\"test_value2\"}"
		assert.Equal(t, expectedQueryStringWithMultipleMetrics, testBuilderWithMultipleMetrics.BuildQuery())
	})
}
