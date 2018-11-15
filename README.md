## mouryou

#### Requirements

  - [Golang](https://golang.org/) >= 1
  - [Apache HTTP Server](http://httpd.apache.org/) with [mod_status](http://httpd.apache.org/docs/2.4/mod/mod_status.html)
  - [IP Virtual Server](http://www.linuxvirtualserver.org/software/ipvs.html)

#### Installation

    $ git clone git://github.com/sai-lab/mouryou.git
    $ cd mouryou
    $ make gom link
    $ make build
    $ sudo bin/mouryou

#### Configuration

`~/.mouryou.json`

```json
{
  "develop_log_level": 0,
  "timeout": 1,
  "sleep": 30,
  "wait": 30,
  "restoration_time": 30,
  "margin": 0.007,
  "algorithm": "BasicSpike",
  "is_weight_change": true,
  "use_hetero": false,
  "adjust_server_num": true,
  "use_operating_ratio": true,
  "use_throughput": false,
  "log_db": "mysql",
  "log_dsb": "root:@tcp(127.0.0.1:3306)/mouryou?parseTime=true",
  "influxdb_addr": "",
  "influxdb_port": "",
  "influxdb_user": "",
  "influxdb_passwd": "",
  "origin_machine_names": [
    "origin-01",
    "origin-02"
  ],
  "always_running_machines": [
    "mirror-01",
    "mirror-02"
  ],
  "start_machine_ids": [
    1,
    2
  ],
  "web_socket": {
    "origin": "http://0.0.0.0/",
    "url": "ws://0.0.0.0:8000/ws"
  },
  "cluster": {
    "load_balancer": {
      "name": "haproxy",
      "virtual_ip": "192.168.11.11",
      "algorithm": "wlc",
      "threshold_out": 0.8,
      "threshold_in": 0.5,
      "margin": 0.05,
      "scale_out": 2,
      "scale_in": 6,
      "diff": 0.2,
      "throughput_algorithm": "MovingAverageV1.2",
      "throughput_moving_average_interval": 3,
      "throughput_scale_in_ratio": 0.1,
      "throughput_scale_out_ratio": 0.5,
      "throughput_scale_out_threshold": 1,
      "throughput_scale_out_time": 5,
      "throughput_scale_in_threshold": 3,
      "throughput_scale_in_time": 5,
      "use_throughput_dynamic_threshold": true,
      "throughput_dynamic_threshold": {
        "0.3": [
          0,
          30
        ],
        "0.5": [
          30,
          60
        ],
        "0.7": [
          60,
          80
        ],
        "0.9": [
          80,
          100
        ]
      }
    },
    "vendors": [
      {
        "name": "azure",
        "virtual_machines": {
          "origin-01": {
            "id": 1,
            "name": "origin-01",
            "host": "192.168.11.01",
            "operation": "booted up",
            "unit_time": 30,
            "unit_cost": 10,
            "throughput_upper_limit": 200,
            "basic_weight": 10,
            "weight": 10
          },
          "origin-02": {
            "id": 2,
            "name": "origin-02",
            "host": "192.168.11.02",
            "operation": "booted up",
            "unit_time": 30,
            "unit_cost": 10,
            "throughput_upper_limit": 200,
            "basic_weight": 10,
            "weight": 10
          },
          "mirror-01": {
            "id": 3,
            "name": "mirror-01",
            "host": "192.168.11.03",
            "operation": "booted up",
            "unit_time": 30,
            "unit_cost": 10,
            "throughput_upper_limit": 200,
            "basic_weight": 10,
            "weight": 10
          },
          "mirror-02": {
            "id": 4,
            "name": "mirror-02",
            "host": "192.168.11.04",
            "operation": "booted up",
            "unit_time": 30,
            "unit_cost": 10,
            "throughput_upper_limit": 200,
            "basic_weight": 10,
            "weight": 10
          },
          "mirror-03": {
            "id": 5,
            "name": "mirror-03",
            "host": "192.168.11.05",
            "operation": "booted up",
            "unit_time": 30,
            "unit_cost": 10,
            "throughput_upper_limit": 200,
            "basic_weight": 10,
            "weight": 10
          }
        }
      }
    ]
  }
}
```

#### License

mouryou is released under the [MIT license](https://raw.githubusercontent.com/hico-horiuchi/mouryou/master/LICENSE).
