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

package entities

import (
	"bytes"
	"fmt"
	"github.com/prometheus/common/model"
	"strings"
)

type Hostname string
type RankedTasks map[Hostname][]Task

// String returns a beautified string representation of ranked tasks.
func (rankedTasks RankedTasks) String() string {
	var buf bytes.Buffer
	const delimiter = "========================================================================"

	for hostname, rankedColocatedTasks := range rankedTasks {
		var info []string
		info = append(info, fmt.Sprintf("HOST = %s", hostname))
		info = append(info, delimiter)
		info = append(info, "\t\t")
		for rank, task := range rankedColocatedTasks {
			info = append(info, fmt.Sprintf("%v, Rank = %d", task, rank))
		}
		info = append(info, delimiter)
		buf.WriteString(strings.Join(info, "\n"))
	}
	return buf.String()
}

type Task struct {
	// TODO decide whether we can get rid of this field.
	Metric model.Metric
	// Weight assigned to the task that was the ranking criteria.
	Weight float64
	// ID represents the unique identifier for the task (Optional).
	ID string
	// Hostname represents the name of the host on which the task was scheduled (Optional).
	Hostname string
}

// GetMetric returns the metric as returned by the query.
func (t Task) GetMetric() model.Metric {
	return t.Metric
}

// GetWeight returns the weight assigned to the task as a result of calibration.
func (t Task) GetWeight() float64 {
	return t.Weight
}

// GetTaskID returns the task identifier.
func (t Task) GetTaskID() string {
	return t.ID
}

// GetHostname returns the name of the host on which the task was scheduled.
func (t Task) GetHostname() string {
	return t.Hostname
}

func (t Task) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	buf.WriteString(fmt.Sprintf("TaskID = %s, Hostname = %s, Weight = %f", t.ID, t.Hostname, t.Weight))
	buf.WriteString("]")
	return buf.String()
}
