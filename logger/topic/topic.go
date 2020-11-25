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

package topic

type Topic int

var (
	Other               = nameToTopic("other")                 // Default topic type.
	Stage               = nameToTopic("stage")                 // Stage in the task ranking process.
	Query               = nameToTopic("query")                 // Query made to Prometheus.
	QueryResult         = nameToTopic("query_result")          // Result of the execution of the query made to Prometheus.
	TaskRankingStrategy = nameToTopic("task_ranking_strategy") // Task ranking strategy.
	TaskRankingResult   = nameToTopic("task_ranking_result")   // Result of applying the configured task ranking strategy.
)

var topicNames []string

func nameToTopic(name string) Topic {
	topicNames = append(topicNames, name)
	// Returning the enumeration value of the topic.
	// This also corresponds to the index (1-indexed) of the associated name.
	return Topic(len(topicNames))
}

func (t Topic) IsValid() bool {
	return (t > 0) && (int(t) <= len(topicNames)) && (t != Other)
}

func (t Topic) String() string {
	return topicNames[t-1]
}

func FromString(s string) Topic {
	switch s {
	case "stage":
		return Stage
	case "query":
		return Query
	case "query_result":
		return QueryResult
	case "task_ranking_strategy":
		return TaskRankingStrategy
	case "task_ranking_result":
		return TaskRankingResult
	default:
		return Other
	}
}
