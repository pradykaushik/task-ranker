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

package prometheus

import (
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/datafetcher"
)

// DataFetcher implements datafetcher.Interface and is used to fetch time series data
// from the given prometheus endpoint.
type DataFetcher struct {
	// The prometheus endpoint that is queried.
	Endpoint string
	// Labels used to filter the time series data fetched from prometheus.
	// Labels []string
	Labels []LabelMatchers
}

type LabelMatchers struct {
	// Label name used to filter the time series data.
	Label string
	// Regex to use for label matching. This is required only when regex based label matching is used.
	Regex string
	// MatchAs is used to specify the operation to use when matching (=, !=, =~, !=~).
	MatchAs LabelMatchOperation
}

// Label match operation to use when filtering time series data.
type LabelMatchOperation int

const (
	Equal         LabelMatchOperation = iota // Equal translates to using '=' operator.
	NotEqual                                 // NotEqual translates to using '!=' operator.
	EqualRegex                               // EqualRegex translates to using '=~' operator.
	NotEqualRegex                            // NotEqualRegex translates to using '!=~' operator.
)

type Option func(f *DataFetcher) error

func NewDataFetcher(options ...Option) (datafetcher.Interface, error) {
	f := new(DataFetcher)
	for _, opt := range options {
		if err := opt(f); err != nil {
			return nil, errors.Wrap(err, "failed to create prometheus data fetcher")
		}
	}
	return f, nil
}

func WithPrometheusEndpoint(endpoint string) Option {
	return func(f *DataFetcher) error {
		if endpoint == "" {
			return errors.New("invalid endpoint")
		}
		f.Endpoint = endpoint
		return nil
	}
}

func WithLabelFilters(labels []LabelMatchers) Option {
	return func(f *DataFetcher) error {
		f.Labels = labels
		return nil
	}
}

func (f *DataFetcher) Fetch() string {
	return "from::fetcher data from prometheus"
}
