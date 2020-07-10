# Task Ranker 
[![GoDoc](https://godoc.org/github.com/pradykaushik/task-ranker?status.svg)](https://godoc.org/github.com/pradykaushik/task-ranker)
![Build](https://github.com/pradykaushik/task-ranker/workflows/Build%20and%20Test/badge.svg)

Rank tasks running as docker containers in a cluster.

Task Ranker runs as a cron job on a specified schedule. Each time the task ranker is run,
it fetches data from [Prometheus](https://prometheus.io/), filters the data as required and then
submits it to a task ranking strategy. The task ranking strategy uses the data received to
calibrate currently running tasks on the cluster and then rank them accordingly. The results
of the strategy are then fed back to the user through callbacks.

You will need to have a [working Golang environment running at least 1.12](https://golang.org/dl/) and a Linux environment.

### How To Use?
Run the below command to download and install Task Ranker.
```commandline
go get github.com/pradykaushik/task-ranker
```

### Environment
Task Ranker can be used in environments where, 
* [Prometheus](https://prometheus.io/) is used to collect container
specific metrics from hosts on the cluster that are running [docker](https://www.docker.com/) containers.
* [cAdvisor](https://github.com/google/cadvisor), a docker native metrics exporter is run on the hosts to export
resource isolation and usage information of running containers.

See [cAdvisor docs](https://github.com/google/cadvisor/blob/master/docs/storage/prometheus.md)
for more information on how to monitor cAdvisor with Prometheus.

#### Container Label Prefixes
CAdvisor [prefixes all container labels with `container_label_`](https://github.com/google/cadvisor/blob/1223982cc4f575354f28f631a3bd00be88ba2f9f/metrics/prometheus.go#L1633).
Given that the Task Ranker only talks to Prometheus, the labels provided should also include these prefixes.
For example, let us say that we launch a task in a docker container using the command below.
```commandline
docker run --label task_id="1234" -t repository/name:version
```
CAdvisor would then export `container_label_task_id` as the container label.

#### Configuration
Task Ranker configuration requires two components to be configured and provided.
1. DataFetcher - Responsible for fetching data from Prometheus, filtering it
    using the provided labels and submitting it to the chosen strategy.
    - Endpoint: [Prometheus HTTP API](https://prometheus.io/docs/prometheus/latest/querying/api/) endpoint.
2. Ranking Strategy - Uses the data to calibrate currently running tasks and then rank them accordingly.
    - Labels: Used for filtering the time series data using the specified [label matching operation](https://prometheus.io/docs/prometheus/latest/querying/basics/).
    - Receiver of the task ranking results.

Task Ranker is configured as shown below.
The below code snippet shows how Task Ranker can be configured to,
* fetch time series data from a Prometheus server running at [http://localhost:9090](http://localhost:9090).
* data is fetched every 5 seconds.
* use the [cpushares](./strategies/taskRankCpuSharesStrategy.go) strategy to rank tasks.
* filter out metrics where `container_label_task_id!=""`.
* filter out metrics where `container_label_task_host!=""`.
* use `container_label_task_id` as the [dedicated label](#dedicated-label-matchers) to help retrieve the task identifier.
* use `container_label_task_host` as the [dedicated label](#dedicated-label-matchers) to help retrieve the hostname on which the task is running.
* use `dummyTaskRanksReceiver` as the receiver of ranked tasks.
```go
type dummyTaskRanksReceiver struct{}

func (r *dummyTaskRanksReceiver) Receive(rankedTasks entities.RankedTasks) {
	log.Println(rankedTasks)
}

prometheusDataFetcher, err = prometheus.NewDataFetcher(
    prometheus.WithPrometheusEndpoint("http://localhost:9090"))

tRanker, err = New(
    WithDataFetcher(prometheusDataFetcher),
    WithSchedule("?/5 * * * * *"),
    WithStrategy("cpushares", []*query.LabelMatcher{
        {Type: query.TaskID, Label: "container_label_task_id", Operator: query.NotEqual, Value: ""},
        {Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"},
    }, new(dummyTaskRanksReceiver), 1*time.Second))
```

You can now also configure the strategies using initialization [options](./strategies/strategy.go). This allows for 
configuring the time duration of range queries, enabling fine-grained control over the number of data points
over which the strategy is applied. See below example for strategy configuration using options.
```go
type dummyTaskRanksReceiver struct{}

func (r *dummyTaskRanksReceiver) Receive(rankedTasks entities.RankedTasks) {
	log.Println(rankedTasks)
}

prometheusDataFetcher, err = prometheus.NewDataFetcher(
    prometheus.WithPrometheusEndpoint("http://localhost:9090"))

tRanker, err = New(
    WithDataFetcher(prometheusDataFetcher),
    WithSchedule("?/5 * * * * *"),
    WithStrategyOptions("cpuutil",
        strategies.WithLabelMatchers([]*query.LabelMatcher{
            {Type: query.TaskID, Label: "container_label_task_id", Operator: query.NotEqual, Value: ""},
            {Type: query.TaskHostname, Label: "container_label_task_host", Operator: query.Equal, Value: "localhost"}}),
        strategies.WithTaskRanksReceiver(new(dummyTaskRanksReceiver)),
        strategies.WithPrometheusScrapeInterval(1*time.Second),
        strategies.WithRange(query.Seconds, 5)))
```

##### Dedicated Label Matchers
Dedicated Label Matchers can be used to retrieve the task ID and host information from data retrieved
from Prometheus. Strategies can mandate the requirement for one or more dedicated labels.

Currently, the following dedicated label matchers are supported.
1. [TaskID](./query/label.go) - This is used to flag a label as one that can be used to fetch the unique identifier of
    a task.
2. [TaskHostname](./query/label.go) - This is used to flag a label as one that can be used to fetch the name of the
    host on which the task is running.
    
Strategies can demand that one or more dedicated labels be provided. For instance, if a strategy
ranks all tasks running on the cluster, then it can mandate only **_TaskID_** dedicated label. On the other
hand if a strategy ranks colocated tasks, then it can mandate both **_TaskID_** and **_TaskHostname_** dedicated labels.

Dedicated label matchers will need to be provided when using strategies that demand them.<br>
The below code snippet shows how a dedicated label can be provided when configuring the Task Ranker.

```go
WithStrategy("strategy-name", []*query.LabelMatcher{
    {Type: query.TaskID, Label: "taskid_label", Operator: query.NotEqual, Value: ""},
    ... // Other label matchers.
})
```

#### Start the Task Ranker
Once the Task Ranker has been configured, then you can start it by calling `tRanker.Start()`.

### Test Locally
#### Setup
Run [`./create_test_env`](./create_test_env) to,
1. bring up a docker-compose installation running Prometheus and cAdvisor.
2. run [tasks](taskdockerfile) in docker containers.

Each container is allocated different cpu-shares.
For more information on running Prometheus and cAdvisor locally see [here](https://prometheus.io/docs/guides/cadvisor/#monitoring-docker-container-metrics-using-cadvisor).

Once you have Prometheus and cAdvisor running (test by running `curl http://localhost:9090/metrics` or use the browser),

#### Test
Now run the below command to run tests.
```commandline
go test -v ./...
```

The task ranking results are displayed on the console.
Below is what it will look like.
```commandline
HOST = localhost
========================================================================
		
[TaskID = <task id>,Hostname = localhost,Weight = <weight>,], Rank = 0
[TaskID = <task id>,Hostname = localhost,Weight = <weight>,], Rank = 1
[TaskID = <task id>,Hostname = localhost,Weight = <weight>,], Rank = 2
...
[TaskID = <task id>,Hostname = localhost,Weight = <weight>,], Rank = n
========================================================================
```

#### Tear-Down
Once finished testing, tear down the test environment by running [`./tear_down_test_env`](./tear_down_test_env).
