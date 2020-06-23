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
	"bytes"
	"fmt"
	"strings"
)

// Builder represents a query builder that is used to build a query string in promQL format.
type Builder struct {
	// The name of the metric to be queried. This is a required field.
	metric string
	// List of labels to be used to filter the data. This is an optional field.
	labelMatchers []*LabelMatcher
	// The unit of time to use when performing range queries. This is an optional field.
	// If this field is not initialized, then
	timeUnit     TimeUnit
	timeDuration uint
	// TODO (pkaushi1) support functions.
}

// Singleton builder instance.
var builderInstance *Builder

// GetBuilder instantiates the singleton Builder instance if required
// and then returns it.
func GetBuilder(options ...Option) *Builder {
	if builderInstance == nil {
		builderInstance = new(Builder)
		for _, opt := range options {
			opt(builderInstance)
		}
	}

	return builderInstance
}

// BuildQuery builds and returns the query string.
func (b Builder) BuildQuery() string {
	var buf bytes.Buffer
	buf.WriteString(b.metric)
	var filters []string
	for _, m := range b.labelMatchers {
		filters = append(filters, m.String())
	}
	buf.WriteString(fmt.Sprintf("{%s}", strings.Join(filters, ",")))
	if b.timeUnit.IsValid() {
		buf.WriteString(fmt.Sprintf("[%d%s]", b.timeDuration, b.timeUnit))
	}
	return buf.String()
}

type Option func(*Builder)

// WithMetric returns an option that initializes the name of the metric to query.
func WithMetric(metric string) Option {
	return func(b *Builder) {
		b.metric = metric
	}
}

// WithLabelMatchers returns an option that initializes the label matchers to use
// as filters in the query string.
func WithLabelMatchers(labelMatchers ...*LabelMatcher) Option {
	return func(b *Builder) {
		b.labelMatchers = append(b.labelMatchers, labelMatchers...)
	}
}

// WithRange returns an option that initializes the time unit and the duration when
// performing range queries.
func WithRange(timeUnit TimeUnit, durationQty uint) Option {
	return func(b *Builder) {
		b.timeUnit = timeUnit
		b.timeDuration = durationQty
	}
}
