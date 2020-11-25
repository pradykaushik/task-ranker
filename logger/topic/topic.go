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
