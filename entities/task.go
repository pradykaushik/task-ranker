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
)

type RankedTask struct {
	Metric model.Metric
	Weight float64
}

// GetMetric returns the metric as returned by the query.
func (t RankedTask) GetMetric() model.Metric {
	return t.Metric
}

// GetWeight returns the weight assigned to the task as a result of calibration.
func (t RankedTask) GetWeight() float64 {
	return t.Weight
}

func (t RankedTask) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Metric: %v\n", t.Metric))
	buf.WriteString(fmt.Sprintf("Weight: %f\n", t.Weight))
	return buf.String()
}
