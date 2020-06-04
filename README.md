# Task Ranker
Rank tasks running as docker containers in a cluster.

### Usage
[Follow the instructions here](https://git-scm.com/book/en/v2/Git-Tools-Submodules) to include this repository as a git submodule.

#### Configure
Task Ranker is configured as shown below.
```go
tRanker, err := New(
    WithPrometheusEndpoint("http://localhost:9090"),
    WithFilterLabelsZeroValues([]string{"label1", "label2"}),
    WithStrategy("cpushares", &dummyTaskRankReceiver{}),
    WithSchedule("?/5 * * * * *"))
```

#### Test
Test the module using the below command.
```commandline
go test -v ./...
```
