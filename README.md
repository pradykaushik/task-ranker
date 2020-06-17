# Task Ranker
Rank tasks running as docker containers in a cluster.

Task Ranker runs as a cron job on a specified schedule. Each time the task ranker is run,
it fetches data from [Prometheus](https://prometheus.io/), filters the data as required and then
submits it to a ranking strategy. The task ranking strategy uses the data received to
calibrate currently running tasks on the cluster and then rank them accordingly. The results
of the strategy are then fed back to the user through callbacks.

You will need to have a [working Golang environment running at least 1.12](https://golang.org/dl/).

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

#### Configure
Task Ranker configuration requires two components to be configured and provided.
1. DataFetcher - Responsible for fetching data from Prometheus, filtering it
    using the provided labels and submitting it to the chosen strategy.
    - Endpoint: [Prometheus HTTP API](https://prometheus.io/docs/prometheus/latest/querying/api/) endpoint.
2. Ranking Strategy - Uses the data to calibrate currently running tasks and then rank them accordingly.
    - Labels: Used for filtering the time series data using the specified [label matching operation](https://prometheus.io/docs/prometheus/latest/querying/basics/).
    - Receiver of the task ranking results.

Task Ranker is configured as shown below.
```go
type dummyTaskRankReceiver struct{}

func (r *dummyTaskRankReceiver) Receive(rankedTasks []entities.Task) {
	log.Println(rankedTasks)
}

prometheusDataFetcher, err = prometheus.NewDataFetcher(
    prometheus.WithPrometheusEndpoint("http://localhost:9090"))

tRanker, err = New(
    WithDataFetcher(prometheusDataFetcher),
    WithSchedule("?/5 * * * * *"),
    WithStrategy("cpushares", []*query.LabelMatcher{
        {Label: "label1", Operator: query.NotEqual, Value: ""},
        {Label: "label2", Operator: query.NotEqual, Value: ""},
    }, &dummyTaskRankReceiver{}))
```

##### Container Label Prefixes
CAdvisor [prefixes all container labels with `container_label_`](https://github.com/google/cadvisor/blob/1223982cc4f575354f28f631a3bd00be88ba2f9f/metrics/prometheus.go#L1633).
Given that the Task Ranker only talks to Prometheus, the labels provided should also include these prefixes.

For example, let us say that we launch a task in a docker container using the command below.
```commandline
docker run --label task_id="1234" -t repository/name:version
```
CAdvisor would then export `container_label_task_id` as the container label.

#### Start the Task Ranker
Once the Task Ranker has been configured, then you can start it by calling `tRanker.Start()`.

### Test Locally
Run `docker-compose up` to bring up a docker-compose installation running Prometheus and cAdvisor.
For more information on running Prometheus and cAdvisor locally see [here](https://prometheus.io/docs/guides/cadvisor/#monitoring-docker-container-metrics-using-cadvisor).

Once you have Prometheus and cAdvisor running (test by running `curl http://localhost:9090/metrics` or using the browser),
run the below command to launch three containers to simulate to tasks.

```commandline
docker run --name test_container_1 --label task_name="test_task_ubuntu_1" --cpu-shares 1024 -it ubuntu:latest /bin/bash
docker run --name test_container_2 --label task_name="test_task_ubuntu_2" --cpu-shares 2048 -it ubuntu:latest /bin/bash
docker run --name test_container_3 --label task_name="test_task_ubuntu_3" --cpu-shares 3072 -it ubuntu:latest /bin/bash
```
Each container is allocated different cpu-shares to ease the verification of task ranking results.

Now run the below command to run tests.
```commandline
go test -v ./...
```

The task ranking results are displayed on the console. Below is what it will look like.
```commandline
Metric: container_spec_cpu_shares{container_label_task_name="test_task_ubuntu_3", id="/docker/<container_id>", image="ubuntu:latest", instance="cadvisor:8080", job="cadvisor", name="test_container_3"}
Weight: 3072.000000

Metric: container_spec_cpu_shares{container_label_task_name="test_task_ubuntu_2", id="/docker/<container_id>", image="ubuntu:latest", instance="cadvisor:8080", job="cadvisor", name="test_container_2"}
Weight: 2048.000000

Metric: container_spec_cpu_shares{container_label_task_name="test_task_ubuntu_1", id="/docker/<container_id>", image="ubuntu:latest", instance="cadvisor:8080", job="cadvisor", name="test_container_1"}
Weight: 1024.000000
```
