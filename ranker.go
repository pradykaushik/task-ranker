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
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/pradykaushik/task-ranker/strategies/factory"
)

type TaskRanker struct {
	PrometheusEndpoint string
	Strategy           strategies.Strategy
	FilterLabels       []string
	Schedule           string
}

func New(options ...Option) (*TaskRanker, error) {
	tRanker := new(TaskRanker)
	for _, opt := range options {
		if err := opt(tRanker); err != nil {
			return nil, errors.Wrap(err, "failed to create task ranker")
		}
	}
	return tRanker, nil
}

type Option func(*TaskRanker) error

func WithPrometheusEndpoint(promEndpoint string) Option {
	return func(tRanker *TaskRanker) error {
		if promEndpoint == "" {
			return errors.New("invalid endpoint")
		}
		tRanker.PrometheusEndpoint = promEndpoint
		return nil
	}
}

func WithStrategy(strategy string, receiver strategies.TaskRanksReceiver) Option {
	return func(tRanker *TaskRanker) error {
		if strategy == "" {
			return errors.New("invalid strategy")
		}

		if s, err := factory.GetTaskRankStrategy(strategy); err != nil {
			return err
		} else {
			tRanker.Strategy = s
			strategies.Build(tRanker.Strategy, receiver)
		}
		return nil
	}
}

func WithSchedule(schedule string) Option {
	return func(tRanker *TaskRanker) error {
		if schedule == "" {
			return errors.New("invalid schedule")
		}
		tRanker.Schedule = schedule
		return nil
	}
}

func WithFilterLabelsZeroValues(labels []string) Option {
	return func(tRanker *TaskRanker) error {
		for _, l := range labels {
			tRanker.FilterLabels = append(tRanker.FilterLabels, l)
		}
		return nil
	}
}
