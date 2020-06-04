/**
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package taskranker

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"os"
)

type TaskRanker struct {
	PrometheusEndpoint string   `yaml:"prometheus_endpoint"`
	Strategy           string   `yaml:"strategy"`
	FilterLabels       []string `yaml:"filter_labels"`
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

func (tRanker *TaskRanker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tempRanker struct {
		PrometheusEndpoint string   `yaml:"prometheus_endpoint"`
		Strategy           string   `yaml:"strategy"`
		FilterLabels       []string `yaml:"filter_labels"`
	}

	t := new(tempRanker)
	if err := unmarshal(t); err != nil {
		return err
	}
	// Initializing members.
	tRanker.PrometheusEndpoint = t.PrometheusEndpoint
	tRanker.Strategy = t.Strategy
	tRanker.FilterLabels = t.FilterLabels
	return nil
}

type Option func(*TaskRanker) error

func WithConfigFile(filename string) Option {
	return func(tRanker *TaskRanker) error {
		file, err := os.Open(filename)
		if err != nil {
			return errors.Wrap(err, "failed to read config file")
		}

		err = yaml.NewDecoder(file).Decode(tRanker)
		if err != nil {
			return errors.Wrap(err, "failed to decode config file")
		}
		return nil
	}
}
