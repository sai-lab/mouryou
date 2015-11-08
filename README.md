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
    $ bin/mouryou

#### Configuration

`~/.mouryou.json`

    {
      "cluster": {
        "load_balancer": {
          "virtual_ip": "192.168.11.11",
          "algorithm": "wlc",
          "threshold": 0.8,
          "margin": 0.05,
          "scale_out": 2,
          "scale_in": 6
        },
        "hypervisors": [
          {
            "host": "192.168.11.20",
            "virtual_machines": [
              {
                "name": "web-server-1",
                "host": "192.168.11.21"
              }
            ]
          }
        ]
      },
      "timeout": 1,
      "sleep": 30,
      "wait": 30,
      "margin": 0.007,
      "web_socket": {
        "origin": "http://localhost/",
        "url": "ws://localhost:8000/ws"
      }
    }

#### License

mouryou is released under the [MIT license](https://raw.githubusercontent.com/hico-horiuchi/mouryou/master/LICENSE).
