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

package query

import (
	"bytes"
	"fmt"
)

// LabelMatcher represents a single label matcher containing the label name (key),
// the matching operator and the value.
type LabelMatcher struct {
	// Type of the label matcher.
	// This information is used internally and is not required to build the prometheus query.
	// If Type != Other, then label matcher is considered a dedicated label matcher.
	Type LabelType
	// Label name used to filter the time series data.
	Label string
	// Operator is used to specify the operator to use when matching (=, !=, =~, !=~).
	Operator LabelMatchOperator
	// Value to match for.
	// If the regex matching, then value represents the regular expression.
	Value string
}

// Label match operator to use when filtering time series data.
type LabelMatchOperator int

// Starting indices from 1 to not have Equal as the default matching operator.
// This also helps checking if a matching operator was provided. If the value is 0,
// then it would indicate that no matching operator was provided.
var (
	Equal         = nameToOperator("=")   // Value = 1. Equal translates to using '=' operator.
	NotEqual      = nameToOperator("!=")  // NotEqual translates to using '!=' operator.
	EqualRegex    = nameToOperator("=~")  // EqualRegex translates to using '=~' operator.
	NotEqualRegex = nameToOperator("!=~") // NotEqualRegex translates to using '!=~' operator.
)

var labelMatchOperatorNames []string

func nameToOperator(name string) LabelMatchOperator {
	labelMatchOperatorNames = append(labelMatchOperatorNames, name)
	// Returning the enumeration value of the operator.
	// This also corresponds to the index (1-indexed) of the associated name.
	return LabelMatchOperator(len(labelMatchOperatorNames))
}

// Type of the label.
type LabelType int

// By default, label matchers are assigned the "Other" type. However, if the label in a label matcher
// can be used to fetch the task identifier and/or name of the host on which the task is currently running,
// the type of the label matcher should be set to TaskID or TaskHostname respectively.
var (
	Other        = labelTypeIota() // Default label type.
	TaskID       = labelTypeIota() // Indicates a dedicated label that can be keyed in to get unique task identifier.
	TaskHostname = labelTypeIota() // Indicates a dedicated label that can be keyed in to get name of host on which task is running.
)

var labelTypes []LabelType

func labelTypeIota() LabelType {
	n := len(labelTypes)
	labelTypes = append(labelTypes, LabelType(n))
	return LabelType(n)
}

// String returns the label, operator and value in the format <Label><Operator>"<Value>".
// This string can directly be used in the query string.
func (m LabelMatcher) String() string {
	var buf bytes.Buffer
	buf.WriteString(m.Label)
	buf.WriteString(m.Operator.String())
	buf.WriteString(fmt.Sprintf("\"%s\"", m.Value))
	return buf.String()
}

// IsValid returns if an operator is valid.
func (o LabelMatchOperator) IsValid() bool {
	return (o > 0) && (o <= LabelMatchOperator(len(labelMatchOperatorNames)))
}

// String returns the string representation of the operator.
func (o LabelMatchOperator) String() string {
	return labelMatchOperatorNames[o-1]
}

// IsValid returns if the label type is valid.
func (t LabelType) IsValid() bool {
	return (t >= 0) && (int(t) < len(labelTypes))
}
