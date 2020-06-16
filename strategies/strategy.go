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
)

type Interface interface {
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
}

// Build the strategy object.
func Build(s Interface, labelMatchers []*query.LabelMatcher, receiver TaskRanksReceiver) error {
	if receiver == nil {
		return errors.New("nil receiver provided")
	}

	s.SetTaskRanksReceiver(receiver)
	if err := s.SetLabelMatchers(labelMatchers); err != nil {
		return errors.Wrap(err, "invalid label matchers for strategy")
	}
	return nil
}
