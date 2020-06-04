# Task Ranker
Rank tasks running as docker containers in a cluster.

### Usage
[Follow the instructions here](https://git-scm.com/book/en/v2/Git-Tools-Submodules) to include this repository as a git submodule.

#### Configure
Task Ranker is configured by providing a config file written in yaml format.
A sample config is shown below.
```yaml
prometheus_endpoint: http://localhost:9090
strategy: cpushare_static
filter_labels:
  - task_id
  - task_hostname
```
In the above sample config, `prometheus_endpoint` refers to the endpoint
of the prometheus server from which the task ranker fetches data.

#### Test
Test the module using the below command.
```commandline
go test -v ./...
```
