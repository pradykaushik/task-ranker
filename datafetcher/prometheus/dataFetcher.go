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
	"context"
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/datafetcher"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"time"
)

// DataFetcher implements datafetcher.Interface and is used to fetch time series data
// from the given prometheus endpoint.
type DataFetcher struct {
	// endpoint of the prometheus HTTP server.
	endpoint string
	// The strategy that is to be applied on the fetched data.
	// DataFetcher uses the strategy to obtain basic information about the query such as
	// metric name, labels for filtering, match operation to perform etc.
	strategy strategies.Interface
}

type Option func(f *DataFetcher) error

// NewDataFetcher constructs a DataFetcher instance by applying all the provided options.
// Returns error if any one of the options fails.
func NewDataFetcher(options ...Option) (datafetcher.Interface, error) {
	f := new(DataFetcher)
	for _, opt := range options {
		if err := opt(f); err != nil {
			return nil, errors.Wrap(err, "failed to create prometheus data fetcher")
		}
	}
	return f, nil
}

// WithPrometheusEndpoint returns an option that initializes the prometheus HTTP server endpoint.
func WithPrometheusEndpoint(endpoint string) Option {
	return func(f *DataFetcher) error {
		if endpoint == "" {
			return errors.New("invalid endpoint")
		}
		f.endpoint = endpoint
		return nil
	}
}

// SetStrategy sets the task ranking strategy.
func (f *DataFetcher) SetStrategy(s strategies.Interface) {
	f.strategy = s
}

// GetEndpoint returns the prometheus HTTP server endpoint.
func (f *DataFetcher) GetEndpoint() string {
	return f.endpoint
}

// Fetch the data from prometheus, filter it using the provided labels, matching operations
// and corresponding values and return the result.
// Time series is fetched using the v1 API. Note that v1 API is still in experimental stage and
// therefore we need to watch out for compatibility issues.
func (f *DataFetcher) Fetch() (result model.Value, err error) {
	queryBuilder := query.GetBuilder(
		query.WithMetric(f.strategy.GetMetric()),
		query.WithLabelMatchers(f.strategy.GetLabelMatchers()...),
		query.WithRange(f.strategy.GetRange()))
	queryString := queryBuilder.BuildQuery()
	// Following examples from here - https://github.com/prometheus/client_golang/blob/master/api/prometheus/v1/example_test.go.
	var client api.Client
	client, err = api.NewClient(api.Config{Address: f.endpoint})
	if err != nil {
		err = errors.Wrap(err, "failed to fetch data from prometheus")
		return
	}
	v1Api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// TODO do not ignore warnings. maybe log them?
	result, _, err = v1Api.Query(ctx, queryString, time.Now())
	return
}
