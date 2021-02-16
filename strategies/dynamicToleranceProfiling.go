package strategies

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/logger"
	"github.com/pradykaushik/task-ranker/logger/topic"
	"github.com/pradykaushik/task-ranker/query"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

// Just logging different metrics.
// Use it to monitor and log metrics pertaining to tasks running on one single host.
type DynamicToleranceProfiler struct {
	// labels used to filter the time series data fetched from prometheus.
	labels []*query.LabelMatcher
	// dedicatedLabelNameTaskID is the dedicated label to use when filtering metrics based on task id.
	// Storing this for quick access to prevent another O(n) search through labels.
	dedicatedLabelNameTaskID model.LabelName
	// dedicatedLabelNameTaskHostname is the dedicated label to use when filtering metrics on a hostname basis.
	// Storing this for quick access to prevent another O(n) search through labels.
	dedicatedLabelNameTaskHostname model.LabelName
	// prometheusScrapeInterval corresponds to the time interval between two successive metrics scrapes.
	prometheusScrapeInterval time.Duration
	// previousTotalCpuUsage stores the sum of cumulative cpu usage seconds for each running task.
	previousTotalCpuUsage map[string]*cpuUsageDataPoint
	// taskMetrics stores the data for different metrics monitored for each task.
	taskMetrics map[string]map[metric]metricData
	// Time duration for range query.
	// Note that there is a caveat in using range queries to retrieve cpu time for containers.
	// If the tasks are pinned, then using a range > 1s works as we would always get the necessary data points for each cpu (thread).
	// On the other hand, if the tasks are not pinned, then there is no guarantee that the necessary number of data points be available
	// as the cpu scheduler can preempt and re-schedule the task on any available cpu.
	// Therefore, to avoid confusion, this strategy does not use the range query.
	// TODO (pkaushi1) get rid of this eventually.
	rangeTimeUnit query.TimeUnit
	rangeQty      uint
}

// metrics monitored per task.
type metricData float64
type metric string
const cpuSharesMetric metric = "cpu-shares"
const cpuUtilMetric metric = "cpu-util-percentage"

const cpuSharesDefaultValue metricData = -1
const cpuUtilDefaultValue metricData = -1

func (s *DynamicToleranceProfiler) Init() {
	s.rangeTimeUnit = query.None
	s.rangeQty = 0
	s.previousTotalCpuUsage = make(map[string]*cpuUsageDataPoint)
	s.taskMetrics = make(map[string]map[metric]metricData)
}

// SetPrometheusScrapeInterval sets the scrape interval of prometheus.
func (s *DynamicToleranceProfiler) SetPrometheusScrapeInterval(prometheusScrapeInterval time.Duration) {
	s.prometheusScrapeInterval = prometheusScrapeInterval
}

// SetTaskRanksReceiver sets the receiver of the task ranking results.
func (s *DynamicToleranceProfiler) SetTaskRanksReceiver(receiver TaskRanksReceiver) {
}

// Execute the strategy using the provided data.
func (s *DynamicToleranceProfiler) Execute(data model.Value) {
	var vector model.Vector
	switch data.Type() {
	case model.ValVector:
		vector = data.(model.Vector)
	default:
		// invalid value type.
		// TODO do not ignore this. maybe log it?
		return
	}

	// Stores the total cumulative cpu usage for each running task.
	var nowTotalCpuUsage = make(map[string]*cpuUsageDataPoint)
	for _, sample := range vector {
		var taskId model.LabelValue
		var ok bool
		var noDedicatedLabelTaskId, noDedicatedLabelHostname bool
		taskId, noDedicatedLabelTaskId = sample.Metric[s.dedicatedLabelNameTaskID]
		_, noDedicatedLabelHostname = sample.Metric[s.dedicatedLabelNameTaskHostname]

		fmt.Printf("%t and %t\n", noDedicatedLabelTaskId, noDedicatedLabelHostname)

		if noDedicatedLabelTaskId && noDedicatedLabelHostname {
			continue // ignore metric.
		}

		fmt.Println("not skipping")

		if _, ok = s.taskMetrics[string(taskId)]; !ok {
			fmt.Println("adding entry for ", taskId)
			s.taskMetrics[string(taskId)] = make(map[metric]metricData)
		}

		switch sample.Metric["__name__"] {
		case "container_spec_cpu_shares":
			fmt.Println("container_spec_cpu_shares = ", sample.Value)
			if _, ok = s.taskMetrics[string(taskId)][cpuSharesMetric]; !ok {
				s.taskMetrics[string(taskId)][cpuSharesMetric] = metricData(sample.Value)
			}

		case "container_cpu_usage_seconds_total":
			if _, ok = nowTotalCpuUsage[string(taskId)]; !ok {
				// Creating entry to record current total cumulative cpu usage.
				nowTotalCpuUsage[string(taskId)] = &cpuUsageDataPoint{
					totalCumulativeCpuUsage: float64(sample.Value),
					timestamp:               sample.Timestamp,
				}
			} else {
				// Adding cumulative cpu usage seconds for task on a cpu.
				nowTotalCpuUsage[string(taskId)].totalCumulativeCpuUsage += float64(sample.Value)
			}
		default:
			// shouldn't be here.
			// unwanted metric.
		}
	}

	// need to record cpu utilization.
	for taskId, dataPoint := range nowTotalCpuUsage {
		var cpuUtil = cpuUtilDefaultValue
		if prevDataPoint, ok := s.previousTotalCpuUsage[taskId]; !ok {
			s.previousTotalCpuUsage[taskId] = dataPoint
		} else {
			cpuUtil = metricData(s.round(s.cpuUtil(
				prevDataPoint.totalCumulativeCpuUsage,
				prevDataPoint.timestamp,
				dataPoint.totalCumulativeCpuUsage,
				dataPoint.timestamp)))
		}

		fmt.Println("cpuutil = ", cpuUtil)
		s.taskMetrics[taskId][cpuUtilMetric] = cpuUtil
	}

	serialized, err := json.Marshal(s.taskMetrics)
	fmt.Println(s.taskMetrics)
	if err == nil {
		if len(s.taskMetrics) > 0 {
			logger.WithFields(logrus.Fields{
				topic.TaskRankingStrategy.String(): "dT-profiling",
				topic.TaskRankingResult.String(): string(serialized),
			}).Log(logrus.InfoLevel)
		}
	}
}

// round the cpu utilization (%) to the above specified precision.
func (s DynamicToleranceProfiler) round(cpuUtil float64) float64 {
	multiplier := math.Pow10(cpuUtilPrecision)
	magnified := cpuUtil * multiplier
	return math.Round(magnified) / multiplier
}

// cpuUtil calculates and returns the cpu utilization (%) for the task.
// The time difference (in seconds) of the two data points is used as the elapsed time. Note that this okay
// as the task ranker schedule (in seconds) is a multiple of the prometheus scrape interval.
func (s DynamicToleranceProfiler) cpuUtil(
	prevTotalCpuUsage float64,
	prevTotalCpuUsageTimestamp model.Time,
	nowTotalCpuUsage float64,
	nowTotalCpuUsageTimestamp model.Time) float64 {

	// timestamps are in milliseconds and therefore dividing by 1000 to convert to seconds.
	timeDiffSeconds := float64(nowTotalCpuUsageTimestamp-prevTotalCpuUsageTimestamp) / 1000
	return 100.0 * ((nowTotalCpuUsage - prevTotalCpuUsage) / timeDiffSeconds)
}

// GetMetrics returns the names of the metrics to query.
func (s DynamicToleranceProfiler) GetMetrics() []string {
	// TODO convert metrics to constants.
	return []string{
		"container_cpu_usage_seconds_total",
		"container_spec_cpu_shares",
	}
}

// SetLabelMatchers sets the label matchers to use when filtering data.
// This strategy mandates that a dedicated label be provided for filtering metrics based on TaskID and Hostname.
func (s *DynamicToleranceProfiler) SetLabelMatchers(labelMatchers []*query.LabelMatcher) error {
	var foundDedicatedLabelMatcherTaskID bool
	var foundDedicatedLabelMatcherTaskHostname bool
	for _, l := range labelMatchers {
		if !l.Type.IsValid() {
			return errors.New("invalid label matcher type")
		} else if l.Type == query.TaskID {
			foundDedicatedLabelMatcherTaskID = true
			s.dedicatedLabelNameTaskID = model.LabelName(l.Label)
		} else if l.Type == query.TaskHostname {
			foundDedicatedLabelMatcherTaskHostname = true
			s.dedicatedLabelNameTaskHostname = model.LabelName(l.Label)
		}
	}

	if !foundDedicatedLabelMatcherTaskID {
		return errors.New("no dedicated task ID label matcher found")
	} else if !foundDedicatedLabelMatcherTaskHostname {
		return errors.New("no dedicated task hostname label matcher found")
	}

	s.labels = labelMatchers
	return nil
}

// GetLabelMatchers returns the label matchers to be used to filter data.
func (s DynamicToleranceProfiler) GetLabelMatchers() []*query.LabelMatcher {
	return s.labels
}

// GetRange returns the time unit and duration for how far back (in seconds) values need to be fetched.
func (s DynamicToleranceProfiler) GetRange() (query.TimeUnit, uint) {
	return s.rangeTimeUnit, s.rangeQty
}

// SetRange sets the time duration for the range query.
// As this strategy does not use range queries a call to this method results in a NO-OP.
func (s *DynamicToleranceProfiler) SetRange(_ query.TimeUnit, _ uint) {}
