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
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/prometheus/common/model"
	"time"
)

type Interface interface {
	// Initialize any other internal data structure and perform any further setup operations.
	Init()
	// SetPrometheusScrapeInterval sets the prometheus scrape interval.
	SetPrometheusScrapeInterval(time.Duration)
	// SetTaskRanksReceiver registers a receiver of the task ranking results.
	// This receiver is a callback and is used to pass the result of applying
	// the strategy to rank tasks.
	SetTaskRanksReceiver(TaskRanksReceiver)
	// Execute the strategy.
	Execute(model.Value)
	// GetMetric returns the metric to pull data for.
	// Note: This has to be a single metric name.
	GetMetric() string
	// SetLabelMatchers sets the label matchers to use to filter data.
	// Strategy implementations can perform additional validations on the provided label matchers.
	SetLabelMatchers([]*query.LabelMatcher) error
	// GetLabelMatchers returns the labels and corresponding matching operators to use
	// filter out data that is not required by this strategy.
	GetLabelMatchers() []*query.LabelMatcher
	// Range returns the duration specifying how far back in time data needs to be fetched.
	// Returns the unit of time along with an integer quantifying the duration.
	GetRange() (query.TimeUnit, uint)
	// SetRange sets the time duration for the range query.
	SetRange(query.TimeUnit, uint)
}

// Build the strategy object.
func Build(s Interface, options ...Option) error {
	s.Init()
	for _, opt := range options {
		if err := opt(s); err != nil {
			return errors.Wrap(err, "failed to build strategy")
		}
	}
	return nil
}

// Options for configuring strategies.
type Option func(Interface) error

// WithLabelMatchers returns an option that initializes the label matchers to be used by the strategy.
func WithLabelMatchers(labelMatchers []*query.LabelMatcher) Option {
	return func(strategy Interface) error {
		if labelMatchers == nil {
			return errors.New("invalid label matchers: nil provided")
		}
		return strategy.SetLabelMatchers(labelMatchers)
	}
}

// WithRange returns an option that initializes the time unit and duration, if using range queries.
func WithRange(timeUnit query.TimeUnit, qty uint) Option {
	return func(strategy Interface) error {
		if !timeUnit.IsValid() {
			return errors.New("invalid time unit provided for range")
		}
		if qty == 0 {
			return errors.New("time duration cannot be 0")
		}
		strategy.SetRange(timeUnit, qty)
		return nil
	}
}

// WithTaskRanksReceiver returns an option that initializes the receiver to which the task ranking results
// are submitted.
func WithTaskRanksReceiver(receiver TaskRanksReceiver) Option {
	return func(strategy Interface) error {
		if receiver == nil {
			return errors.New("nil receiver provided")
		}
		strategy.SetTaskRanksReceiver(receiver)
		return nil
	}
}

// WithPrometheusScrapeInterval returns an option that initializes the prometheus scrape interval.
func WithPrometheusScrapeInterval(prometheusScrapeInterval time.Duration) Option {
	return func(strategy Interface) error {
		strategy.SetPrometheusScrapeInterval(prometheusScrapeInterval)
		return nil
	}
}
