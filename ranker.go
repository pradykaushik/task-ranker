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
	df "github.com/pradykaushik/task-ranker/datafetcher"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/pradykaushik/task-ranker/strategies/factory"
	"github.com/pradykaushik/task-ranker/util"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

// TaskRanker fetches data pertaining to currently running tasks, deploys a strategy
// to rank them and then feeds the results back to the caller.
// Runs as a cron job on the defined schedule.
type TaskRanker struct {
	// DataFetcher used to pull task/container specific data.
	DataFetcher df.Interface
	// Strategy to use for calibration and ranking of tasks using the data fetched.
	Strategy strategies.Interface
	// Schedule on which the ranker runs. The schedule should follow the cron schedule format.
	// See https://en.wikipedia.org/wiki/Cron.
	// Alternatively, Seconds can also be specified as part of the schedule.
	// See https://godoc.org/github.com/robfig/cron.
	Schedule cron.Schedule
	// runner is a cron job runner used to run the task ranker on the specified schedule.
	runner *cron.Cron
	// termCh is a channel used to signal the task ranker to stop.
	termCh *util.SignalChannel
}

func New(options ...Option) (*TaskRanker, error) {
	tRanker := new(TaskRanker)
	for _, opt := range options {
		if err := opt(tRanker); err != nil {
			return nil, errors.Wrap(err, "failed to create task ranker")
		}
	}
	tRanker.termCh = util.NewSignalChannel()
	return tRanker, nil
}

type Option func(*TaskRanker) error

func WithDataFetcher(dataFetcher df.Interface) Option {
	return func(tRanker *TaskRanker) error {
		if dataFetcher == nil {
			return errors.New("invalid data fetcher")
		}
		tRanker.DataFetcher = dataFetcher
		return nil
	}
}

func WithStrategy(
	strategy string,
	labelMatchers []*query.LabelMatcher,
	receiver strategies.TaskRanksReceiver,
	prometheusScrapeInterval time.Duration) Option {

	return func(tRanker *TaskRanker) error {
		if strategy == "" {
			return errors.New("invalid strategy")
		}

		// TODO validate arguments.
		if s, err := factory.GetTaskRankStrategy(strategy); err != nil {
			return err
		} else {
			tRanker.Strategy = s
			err := strategies.Build(tRanker.Strategy, labelMatchers, receiver, prometheusScrapeInterval)
			if err != nil {
				return errors.Wrap(err, "failed to build strategy")
			}
			tRanker.DataFetcher.SetStrategy(s)
		}
		return nil
	}
}

func WithSchedule(specString string) Option {
	return func(tRanker *TaskRanker) error {
		schedule, err := cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(specString)
		if err != nil {
			return errors.Wrap(err, "invalid schedule")
		}
		tRanker.Schedule = schedule
		return nil
	}
}

func (tRanker *TaskRanker) Start() {
	tRanker.runner = cron.New(cron.WithSeconds())
	tRanker.runner.Schedule(tRanker.Schedule, tRanker)
	tRanker.runner.Start()
}

func (tRanker *TaskRanker) Run() {
	if tRanker.termCh.IsClosed() {
		return
	}
	result, err := tRanker.DataFetcher.Fetch()
	if err != nil {
		log.Println(err.Error())
	} else {
		tRanker.Strategy.Execute(result)
	}
}

func (tRanker *TaskRanker) Stop() {
	tRanker.termCh.Close()
	tRanker.runner.Stop()
}
