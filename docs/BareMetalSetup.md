# Bare Metal Setup of Prometheus and cAdvisor

## cAdvisor Configuration
[cAdvisor](https://github.com/google/cadvisor) needs to be running on each host on which you want to monitor running containers.
We are going to setup cAdvisor as a systemd service.

Follow instructions [here](https://github.com/google/cadvisor/blob/master/docs/development/build.md) to build cAdvisor from source.
Once built, copy the executable into _/usr/local/bin_.

Create a systemd file named _cadvisor.service_ with the below contents in _/etc/systemd/system/_.

**/etc/systemd/system/cadvisor.service**
```shell
[Unit]
Description=cAdvisor
Wants=network-online.target
After=network-online.target docker.service

[Service]
Type=simple
ExecStart=/usr/local/bin/cadvisor -port 9090 -store_container_labels=false -whitelisted_container_labels=LABEL, [LABEL, ...] -docker_only=true

[Install]
WantedBy=multi-user.target
```
1. cAdvisor is being configured to run with root privileges to be able to whitelist docker labels.
2. `-docker\_only=true` tells cAdvisor to report only docker container metrics.
3. [-whitelisted\_container\_labels](https://github.com/google/cadvisor/blob/90f391fddf71801f76f408d8ed191ddc006df323/cmd/cadvisor.go#L72) are a comma separated list of container labels that you assign to containers that need to be monitored. You can then provide these labels are dedicated labels to Task Ranker that will use them as labels in the Prometheus query string.

Run the below command on each host to launch cAdvisor.
```commandline
sudo systemctl start cadvisor
```

## Prometheus Configuration

Create a directory called _prometheus_ in _/etc_.
On the host machine from which you will be collecting container metrics, store the below Prometheus configuration file in _/etc/prometheus/prometheus.yml_.
Checkout [Prometheus configuration docs](https://github.com/prometheus/prometheus/blob/master/docs/configuration/configuration.md) to get more help on configuring Prometheus.
```shell
scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 1s
    static_configs:
      - targets: ['localhost:9090']
  - job_name: '<host1>-cadvisor'
    scrape_interval: 1s
    static_configs:
      - targets: ['<host1-ip>:9090']
  - job_name: '<host2>-cadvisor'
    scrape_interval: 1s
    static_configs:
      - targets: ['<host2-ip>:9090']
  ...
```

Use Prometheus [Makefile targets](https://github.com/prometheus/prometheus/blob/master/Makefile) to build from source.<br>
Run the below command to create the directory in which Prometheus stores its time series database.
```commandline
mkdir /var/lib/prometheus
```
We will now setup Prometheus as a systemd service. Create a systemd file named _prometheus.service_ with the below contents in _/etc/systemd/system_.

**/etc/systemd/system/prometheus.service**
```shell
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
ExecStart=/usr/local/bin/prometheus \
        --config.file /etc/prometheus/prometheus.yml \
        --storage.tsdb.path /var/lib/prometheus/ \
        --web.console.templates=/etc/prometheus/consoles \
        --web.console.libraries=/etc/prometheus/console_libraries

[Install]
WantedBy=multi-user.target
```

Run the below command on each host to launch prometheus.
```commandline
sudo systemctl start prometheus
```

You should now be able to hit [http://localhost:9090](http://localhost:9090) to see the Prometheus UI.
