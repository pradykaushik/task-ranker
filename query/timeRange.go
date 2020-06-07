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

type TimeUnit int

// Starting indices from 1 to not have Seconds as the default time unit.
// This also helps checking if a time unit was provided. If the value is 0, then
// it would indicate that time unit was not provided.
var (
	Seconds = nameToUnit("s") // Value = 1
	Minutes = nameToUnit("m")
	Hours   = nameToUnit("h")
	Days    = nameToUnit("d")
	Weeks   = nameToUnit("w")
	Years   = nameToUnit("y")
)

// timeUnitNames stores the string representation of the time unit that is used in range queries.
var timeUnitNames []string

// Return the string representation of the time unit used in range queries.
func (t TimeUnit) String() string {
	return timeUnitNames[t-1]
}

// IsValid returns if a time unit is valid.
func (t TimeUnit) IsValid() bool {
	return (t > 0) && (t <= TimeUnit(len(timeUnitNames)))
}

// nameToUnit converts the provided time unit name to an integer.
func nameToUnit(name string) TimeUnit {
	timeUnitNames = append(timeUnitNames, name)
	// Returning the enumeration value of the time unit.
	// This also corresponds to the index (1-indexed) of the associated name.
	return TimeUnit(len(timeUnitNames))
}
