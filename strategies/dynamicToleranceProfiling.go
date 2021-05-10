package strategies

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/pradykaushik/task-ranker/entities"
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
	// receiver of the results of task ranking.
	receiver TaskRanksReceiver
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
	// previousInstructionsRetired stores the instructions retired in the previous monitoring cycle.
	// Expects that perf metrics from cAdvisor are aggregated (--disable_metrics="percpu").
	previousPerfMetrics map[string]map[perfMetricName]*perfDataPoint
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

type perfDataPoint struct {
	value     float64
	timestamp model.Time
}

// metrics monitored per task.
type metricData float64
type metric string

// cAdvisor metrics.
const cpuSharesMetric metric = "container_spec_cpu_shares"
const cpuQuotaMetric metric = "container_spec_cpu_quota"
const cpuPeriodMetric metric = "container_spec_cpu_period"
const cpuCfsThrottledPeriodsTotalMetric metric = "container_cpu_cfs_throttled_periods_total"
const cpuCfsThrottledSecondsTotalMetric metric = "container_cpu_cfs_throttled_seconds_total"
const cpuUsageSecondsTotalMetric metric = "container_cpu_usage_seconds_total"
const cpuSchedStatRunQueueSecondsTotalMetric metric = "container_cpu_schedstat_runqueue_seconds_total"
const cpuSchedStatRunPeriodsTotalMetric metric = "container_cpu_schedstat_run_periods_total"
const cpuSchedStatRunSecondsTotalMetric metric = "container_cpu_schedstat_run_seconds_total"
const fsUsageBytesMetric metric = "container_fs_usage_bytes"
const processesMetric metric = "container_processes"
const perfEventsTotalMetric metric = "container_perf_events_total"

// perf event names.
type perfMetricName model.LabelValue

const perfEventNameInstructionsRetired perfMetricName = "instructions_retired"
const perfEventNameCycles perfMetricName = "cycles"

// Derived metrics.
const cpuUtilMetric metric = "cpu_util_percentage"
const cpuUtilDefaultValue metricData = -1
const instructionRetirementRateMetric metric = "instruction_retirement_rate"
const instructionRetirementRateDefaultValue metricData = -1
const cyclesPerSecondMetric metric = "cycles_per_second"
const cyclesPerSecondDefaultValue metricData = -1
const cyclesPerInstructionMetric metric = "cycles_per_instruction"
const cyclesPerInstructionDefaultValue metricData = -1

func (s *DynamicToleranceProfiler) Init() {
	s.rangeTimeUnit = query.None
	s.rangeQty = 0
	s.previousTotalCpuUsage = make(map[string]*cpuUsageDataPoint)
	s.previousPerfMetrics = make(map[string]map[perfMetricName]*perfDataPoint)
	s.taskMetrics = make(map[string]map[metric]metricData)
}

// SetPrometheusScrapeInterval sets the scrape interval of prometheus.
func (s *DynamicToleranceProfiler) SetPrometheusScrapeInterval(prometheusScrapeInterval time.Duration) {
	s.prometheusScrapeInterval = prometheusScrapeInterval
}

// SetTaskRanksReceiver sets the receiver of the task ranking results.
func (s *DynamicToleranceProfiler) SetTaskRanksReceiver(receiver TaskRanksReceiver) {
	s.receiver = receiver
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
	// Stores the number of instructions retired since the last cycle.
	var nowInstructionsRetired = make(map[string]*perfDataPoint)
	var nowCycles = make(map[string]*perfDataPoint)
	// Boolean indicating whether a task has completed.
	var terminatedTasks = make(map[string]struct{})
	for _, sample := range vector {
		var taskId model.LabelValue
		var ok bool
		var foundDedicatedLabelTaskId, foundDedicatedLabelHostname bool
		taskId, foundDedicatedLabelTaskId = sample.Metric[s.dedicatedLabelNameTaskID]
		_, foundDedicatedLabelHostname = sample.Metric[s.dedicatedLabelNameTaskHostname]

		if !foundDedicatedLabelTaskId && !foundDedicatedLabelHostname {
			continue // ignore metric.
		}

		// Assuming that the task has terminated.
		// Task will be removed from map if found to still be running.
		terminatedTasks[string(taskId)] = struct{}{}

		if _, ok = s.taskMetrics[string(taskId)]; !ok {
			s.taskMetrics[string(taskId)] = make(map[metric]metricData)
		}

		m := metric(sample.Metric["__name__"])
		// If we find even a single spec metric, then we can conclude that the task has terminated.
		// The reason we are having to do this is because metrics continue to persist for some time after
		// 	a task has terminated. See https://github.com/google/cadvisor/issues/2812.
		if (m == cpuSharesMetric) || (m == cpuQuotaMetric) || (m == cpuPeriodMetric) {
			delete(terminatedTasks, string(taskId))
		}
		switch m {
		case cpuSharesMetric, cpuQuotaMetric, cpuPeriodMetric,
			cpuCfsThrottledPeriodsTotalMetric, cpuCfsThrottledSecondsTotalMetric,
			fsUsageBytesMetric, cpuSchedStatRunQueueSecondsTotalMetric, cpuSchedStatRunPeriodsTotalMetric,
			cpuSchedStatRunSecondsTotalMetric, processesMetric:
			s.taskMetrics[string(taskId)][m] = metricData(sample.Value)

		case cpuUsageSecondsTotalMetric:
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

		case perfEventsTotalMetric:
			// Recording instructions retired.
			var event model.LabelValue
			event, ok = sample.Metric["event"]
			if ok {
				if event == model.LabelValue(perfEventNameInstructionsRetired) {
					if _, ok := nowInstructionsRetired[string(taskId)]; !ok {
						nowInstructionsRetired[string(taskId)] = &perfDataPoint{
							value:     float64(sample.Value),
							timestamp: sample.Timestamp,
						}
					}
				} else if event == model.LabelValue(perfEventNameCycles) {
					if _, ok := nowCycles[string(taskId)]; !ok {
						nowCycles[string(taskId)] = &perfDataPoint{
							value:     float64(sample.Value),
							timestamp: sample.Timestamp,
						}
					}
				}
			}

		default:
			// shouldn't be here.
			// unwanted metric.
		}
	}

	// Remove entries for terminated tasks.
	for taskId := range terminatedTasks {
		delete(nowTotalCpuUsage, taskId)
		delete(nowInstructionsRetired, taskId)
		delete(nowCycles, taskId)
		delete(s.previousPerfMetrics, taskId)
		delete(s.taskMetrics, taskId)
	}

	// calculate and record cpu utilization.
	for taskId, dataPoint := range nowTotalCpuUsage {
		var cpuUtil = cpuUtilDefaultValue
		if prevDataPoint, ok := s.previousTotalCpuUsage[taskId]; !ok {
			s.previousTotalCpuUsage[taskId] = dataPoint
		} else {
			cpuUtil = metricData(s.cpuUtil(
				prevDataPoint.totalCumulativeCpuUsage,
				prevDataPoint.timestamp,
				dataPoint.totalCumulativeCpuUsage,
				dataPoint.timestamp))
		}

		s.taskMetrics[taskId][cpuUtilMetric] = cpuUtil
	}

	// calculate and record instruction retirement rate.
	for taskId, nowIR := range nowInstructionsRetired {
		if _, ok := s.previousPerfMetrics[taskId]; !ok {
			s.previousPerfMetrics[taskId] = make(map[perfMetricName]*perfDataPoint)
		}
		if previousIR, ok := s.previousPerfMetrics[taskId][perfEventNameInstructionsRetired]; !ok {
			// first time recording instructions retired for this task.
			s.taskMetrics[taskId][instructionRetirementRateMetric] = instructionRetirementRateDefaultValue
		} else {
			// IRR = #instructions_retired / elapsed_time_in_seconds.
			s.taskMetrics[taskId][instructionRetirementRateMetric] = metricData(s.rate(
				previousIR.value,
				previousIR.timestamp,
				nowIR.value,
				nowIR.timestamp))
		}
		s.previousPerfMetrics[taskId][perfEventNameInstructionsRetired] = nowIR
	}

	// calculate and record cycles per second.
	for taskId, cycles := range nowCycles {
		if _, ok := s.previousPerfMetrics[taskId]; !ok {
			s.previousPerfMetrics[taskId] = make(map[perfMetricName]*perfDataPoint)
		}
		if previousCycles, ok := s.previousPerfMetrics[taskId][perfEventNameCycles]; !ok {
			// first time recording cycles for this task.
			s.taskMetrics[taskId][cyclesPerSecondMetric] = cyclesPerSecondDefaultValue
		} else {
			s.taskMetrics[taskId][cyclesPerSecondMetric] = metricData(s.rate(
				previousCycles.value,
				previousCycles.timestamp,
				cycles.value,
				cycles.timestamp))
		}
		s.previousPerfMetrics[taskId][perfEventNameCycles] = cycles
	}

	// calculating cycles per instruction (CPI).
	// at this point, for every taskId we have recorded both IPS (IRR) and CPS.
	for _, metrics := range s.taskMetrics {
		instructionRetirementRate, irrOk := metrics[instructionRetirementRateMetric]
		cyclesPerSecond, cpsOk := metrics[cyclesPerSecondMetric]
		if (!irrOk || !cpsOk) ||
			((instructionRetirementRate == instructionRetirementRateDefaultValue) &&
				(cyclesPerSecond == cyclesPerSecondDefaultValue)) {
			metrics[cyclesPerInstructionMetric] = cyclesPerInstructionDefaultValue
		} else {
			metrics[cyclesPerInstructionMetric] = metricData(s.round(float64(cyclesPerSecond/instructionRetirementRate), 2))
		}
	}

	// creating results to send back to receiver.
	// just for dynamic-tolerance purpose, results will look like as shown below.
	// {taskId => [{ID: <metric_name>, Weight: <value>} ... ]}
	result := make(entities.RankedTasks)
	for taskId, metrics := range s.taskMetrics {
		var fields = logrus.Fields{topic.Metrics.String(): "dT-profiler-metrics"}
		fields["taskId"] = taskId
		// this line looks weird as hell. Note that it's just a hack and will not be pushed to master.
		result[entities.Hostname(taskId)] = make([]entities.Task, len(metrics))
		for header, value := range metrics {
			fields[string(header)] = fmt.Sprintf("%.3f", value)
			result[entities.Hostname(taskId)] = append(result[entities.Hostname(taskId)], entities.Task{
				ID:     string(header),
				Weight: float64(value),
			})
		}
		logger.WithFields(fields).Log(logrus.InfoLevel)
	}

	if len(s.taskMetrics) > 0 {
		s.receiver.Receive(result)
	}
}

// Round value to provided precision.
func (s DynamicToleranceProfiler) round(value float64, precision int) float64 {
	multiplier := math.Pow10(precision)
	magnified := value * multiplier
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

func (s DynamicToleranceProfiler) rate(
	previousValue float64,
	previousTimestamp model.Time,
	nowValue float64,
	nowTimestamp model.Time) float64 {
	// timestamps are in milliseconds and therefore dividing by 1000 to convert to seconds.
	timeDiffSeconds := float64(nowTimestamp-previousTimestamp) / 1000
	valueDiff := nowValue - previousValue
	return valueDiff / timeDiffSeconds
}

// GetMetrics returns the names of the metrics to query.
func (s DynamicToleranceProfiler) GetMetrics() []string {
	// TODO convert metrics to constants.
	return []string{
		string(cpuSharesMetric),
		string(cpuQuotaMetric),
		string(cpuPeriodMetric),
		string(cpuCfsThrottledPeriodsTotalMetric),
		string(cpuCfsThrottledSecondsTotalMetric),
		string(cpuUsageSecondsTotalMetric),
		string(cpuSchedStatRunQueueSecondsTotalMetric),
		string(cpuSchedStatRunPeriodsTotalMetric),
		string(cpuSchedStatRunSecondsTotalMetric),
		string(fsUsageBytesMetric),
		string(processesMetric),
		string(perfEventsTotalMetric),
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
	// Adding additional labels for scraping perf metrics.
	s.labels = append(s.labels, &query.LabelMatcher{
		Label:    "event",
		Operator: query.EqualRegex,
		// Hack: using {event=~"<event_name>|"} instead of {event="<event_name>"} to prevent
		// filtering out other metrics for which there would be no data for 'event' label.
		Value: string(perfEventNameInstructionsRetired) + "|" + string(perfEventNameCycles) + "|",
	})
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
