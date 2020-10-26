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
	"fmt"
	"github.com/pkg/errors"
	df "github.com/pradykaushik/task-ranker/datafetcher"
	"github.com/pradykaushik/task-ranker/logger"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/pradykaushik/task-ranker/strategies"
	"github.com/pradykaushik/task-ranker/strategies/factory"
	"github.com/pradykaushik/task-ranker/util"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
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
	// prometheusScrapeInterval corresponds to the time interval between two successive metrics scrapes.
	prometheusScrapeInterval time.Duration
}

func New(options ...Option) (*TaskRanker, error) {
	tRanker := new(TaskRanker)
	for _, opt := range options {
		if err := opt(tRanker); err != nil {
			return nil, errors.Wrap(err, "failed to create task ranker")
		}
	}

	// checking if schedule provided.
	if tRanker.Schedule == nil {
		return nil, errors.New("invalid schedule provided for task ranker")
	}

	// validate task ranker schedule to be a multiple of prometheus scrape interval.
	now := time.Unix(0, 0)
	nextTimeTRankerSchedule := tRanker.Schedule.Next(now)
	tRankerScheduleIntervalSeconds := int(nextTimeTRankerSchedule.Sub(now).Seconds())
	if (tRankerScheduleIntervalSeconds < int(tRanker.prometheusScrapeInterval.Seconds())) ||
		((tRankerScheduleIntervalSeconds % int(tRanker.prometheusScrapeInterval.Seconds())) != 0) {
		return nil, errors.New(fmt.Sprintf("task ranker schedule (%d seconds) should be a multiple of "+
			"prometheus scrape interval (%d seconds)", tRankerScheduleIntervalSeconds, int(tRanker.prometheusScrapeInterval.Seconds())))
	}
	// Providing the prometheus scrape interval to the strategy.
	tRanker.Strategy.SetPrometheusScrapeInterval(tRanker.prometheusScrapeInterval)
	tRanker.termCh = util.NewSignalChannel()

	// Configuring logger.
	err := logger.Configure()
	if err != nil {
		err = errors.Wrap(err, "failed to configure logger")
		if err = logger.Done(); err != nil {
			err = errors.Wrap(err, "failed to shutdown logger")
		}
	}
	return tRanker, err
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

// WithPrometheusScrapeInterval returns an option that initializes the prometheus scrape interval.
func WithPrometheusScrapeInterval(prometheusScrapeInterval time.Duration) Option {
	return func(tRanker *TaskRanker) error {
		if prometheusScrapeInterval == 0 {
			return errors.New("invalid prometheus scrape interval: should be > 0")
		}
		tRanker.prometheusScrapeInterval = prometheusScrapeInterval
		return nil
	}
}

// WithStrategy builds the task ranking strategy associated with the given name using the provided information.
// For backwards compatibility, strategies that use range queries will use the default duration. If the time
// duration for the range query needs to be configured, then use WithStrategyOptions(...) to configure the strategy
// and provide the WithRange(...) option.
func WithStrategy(
	strategy string,
	labelMatchers []*query.LabelMatcher,
	receiver strategies.TaskRanksReceiver) Option {

	return func(tRanker *TaskRanker) error {
		if strategy == "" {
			return errors.New("invalid strategy")
		}

		if s, err := factory.GetTaskRankStrategy(strategy); err != nil {
			return err
		} else {
			tRanker.Strategy = s
			err := strategies.Build(s,
				strategies.WithLabelMatchers(labelMatchers),
				strategies.WithTaskRanksReceiver(receiver))
			if err != nil {
				return errors.Wrap(err, "failed to build strategy")
			}
			tRanker.DataFetcher.SetStrategy(s)
		}
		return nil
	}
}

// WithStrategyOptions builds the strategy associated with the given name using the provided initialization options.
func WithStrategyOptions(strategy string, strategyOptions ...strategies.Option) Option {
	return func(tRanker *TaskRanker) error {
		if strategy == "" {
			return errors.New("invalid strategy")
		}

		if s, err := factory.GetTaskRankStrategy(strategy); err != nil {
			return err
		} else {
			tRanker.Strategy = s
			err := strategies.Build(s, strategyOptions...)
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
	logger.WithFields(logrus.Fields{
		"stage": "task-ranker",
	}).Log(logrus.InfoLevel, "starting task ranker cron job")
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
		logger.WithFields(logrus.Fields{
			"stage": "data-fetcher",
		}).Log(logrus.ErrorLevel, err.Error())
	} else {
		tRanker.Strategy.Execute(result)
	}
}

func (tRanker *TaskRanker) Stop() {
	logger.WithFields(logrus.Fields{
		"stage": "task-ranker",
	}).Log(logrus.InfoLevel, "stopping task ranker cron job")
	tRanker.termCh.Close()
	tRanker.runner.Stop()
	err := logger.Done()
	if err != nil {
		fmt.Printf("failed to shutdown logger: %v", err)
	}
}
